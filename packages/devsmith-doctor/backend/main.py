"""
DevSmith Doctor - Main FastAPI Service

Provides API endpoints for diagnosing and fixing Docker/nginx configuration issues.
Integrates with docker-validate.sh and external tools like nginxfmt and hadolint.
"""

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import List, Optional, Dict, Any
from pathlib import Path
import json
import logging

from doctor.analyzer import IssueAnalyzer
from doctor.fixer import FixGenerator, FixExecutor
from doctor.integrations import ToolIntegrations
from doctor.logger import DoctorLogger

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="DevSmith Doctor",
    description="Intelligent Docker & nginx Configuration Auto-Fixer",
    version="1.0.0"
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3000", "http://localhost:5173"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Initialize components
project_root = Path("/app")
analyzer = IssueAnalyzer(project_root)
fix_generator = FixGenerator(project_root)
fix_executor = FixExecutor(project_root)
integrations = ToolIntegrations(project_root)
doctor_logger = DoctorLogger()


# Pydantic models
class Issue(BaseModel):
    id: str
    type: str
    service: str
    severity: str
    message: str
    details: Optional[Dict[str, Any]] = None


class Fix(BaseModel):
    id: str
    issue_id: str
    commands: List[str]
    explanation: str
    safe_to_auto_apply: bool
    confidence: float
    tool_integrations: Optional[List[str]] = None


class DiagnosisResponse(BaseModel):
    status: str
    total_issues: int
    high_priority: int
    issues_with_fixes: List[Dict[str, Any]]
    timestamp: str


class FixRequest(BaseModel):
    fix_id: str
    auto_apply: bool = False
    dry_run: bool = False


class FixResponse(BaseModel):
    success: bool
    message: str
    output: Optional[str] = None
    errors: Optional[List[str]] = None


# Health check
@app.get("/health")
async def health():
    """Health check endpoint"""
    return {"status": "healthy", "service": "devsmith-doctor"}


@app.get("/api/health")
async def api_health():
    """API health check"""
    return {"status": "healthy"}


# Main endpoints
@app.post("/api/diagnose", response_model=DiagnosisResponse)
async def diagnose():
    """
    Analyze validation status and generate fixes for all issues.
    
    Reads .validation/status.json from docker-validate.sh and generates
    context-aware fixes for each issue found.
    """
    try:
        # Read validation status
        validation_data = analyzer.read_validation_status()
        if not validation_data:
            raise HTTPException(
                status_code=404,
                detail="Validation file not found. Run docker-validate.sh first."
            )
        
        # Parse issues
        issues = analyzer.parse_issues(validation_data)
        
        if not issues:
            return DiagnosisResponse(
                status="healthy",
                total_issues=0,
                high_priority=0,
                issues_with_fixes=[],
                timestamp=analyzer.get_timestamp()
            )
        
        # Generate fixes for each issue
        issues_with_fixes = []
        high_priority = 0
        
        for issue in issues:
            if issue.severity == "high":
                high_priority += 1
            
            # Generate fix
            fix = fix_generator.generate_fix(issue)
            
            # Check if we can use tool integrations
            tool_suggestions = integrations.suggest_tools(issue)
            if tool_suggestions:
                fix.tool_integrations = tool_suggestions
            
            issues_with_fixes.append({
                "issue": issue.dict(),
                "fix": fix.dict()
            })
        
        # Log diagnosis
        doctor_logger.log_diagnosis(len(issues), high_priority)
        
        return DiagnosisResponse(
            status="issues_found",
            total_issues=len(issues),
            high_priority=high_priority,
            issues_with_fixes=issues_with_fixes,
            timestamp=analyzer.get_timestamp()
        )
    
    except Exception as e:
        logger.error(f"Error during diagnosis: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/api/fix", response_model=FixResponse)
async def apply_fix(request: FixRequest):
    """
    Apply a specific fix.
    
    Can run in dry-run mode to preview commands without executing,
    or auto-apply mode to execute immediately.
    """
    try:
        # Get the fix from stored diagnosis
        fix = fix_executor.get_fix_by_id(request.fix_id)
        if not fix:
            raise HTTPException(status_code=404, detail=f"Fix {request.fix_id} not found")
        
        # Dry run - just return what would be executed
        if request.dry_run:
            return FixResponse(
                success=True,
                message="Dry run - no commands executed",
                output="\n".join(fix.commands)
            )
        
        # Execute fix
        if request.auto_apply:
            success, output, errors = fix_executor.execute_fix(fix)
            
            # Log the fix application
            doctor_logger.log_fix_applied(fix, success)
            
            return FixResponse(
                success=success,
                message="Fix applied successfully" if success else "Fix failed",
                output=output,
                errors=errors
            )
        else:
            return FixResponse(
                success=False,
                message="auto_apply must be true to execute fix",
                output=None
            )
    
    except Exception as e:
        logger.error(f"Error applying fix: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/api/fix/batch")
async def apply_fixes_batch(fix_ids: List[str], auto_apply: bool = False):
    """
    Apply multiple fixes in sequence.
    
    Useful for auto-fixing all safe issues at once.
    """
    results = []
    
    for fix_id in fix_ids:
        try:
            result = await apply_fix(FixRequest(fix_id=fix_id, auto_apply=auto_apply))
            results.append({
                "fix_id": fix_id,
                "success": result.success,
                "message": result.message
            })
        except Exception as e:
            results.append({
                "fix_id": fix_id,
                "success": False,
                "message": str(e)
            })
    
    return {
        "total": len(fix_ids),
        "successful": sum(1 for r in results if r["success"]),
        "failed": sum(1 for r in results if not r["success"]),
        "results": results
    }


@app.get("/api/status")
async def get_status():
    """
    Get current system health status.
    
    Returns overall health and any active issues.
    """
    try:
        validation_data = analyzer.read_validation_status()
        if not validation_data:
            return {
                "status": "unknown",
                "message": "No validation data available"
            }
        
        validation = validation_data.get("validation", {})
        status = validation.get("status", "unknown")
        issues = validation.get("issues", [])
        
        return {
            "status": status,
            "total_issues": len(issues),
            "high_priority_issues": sum(1 for i in issues if i.get("severity") == "high"),
            "last_validation": validation.get("timestamp"),
            "services_checked": validation.get("services_checked", 0)
        }
    
    except Exception as e:
        logger.error(f"Error getting status: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/api/patterns")
async def get_patterns():
    """
    Get all known fix patterns.
    
    Returns the pattern library for reference or debugging.
    """
    return fix_generator.get_all_patterns()


@app.post("/api/integrate/nginxfmt")
async def run_nginxfmt():
    """
    Run nginxfmt on nginx.conf to auto-format and fix common issues.
    """
    try:
        success, output = integrations.run_nginxfmt()
        return {
            "success": success,
            "output": output,
            "message": "nginx.conf formatted successfully" if success else "nginxfmt failed"
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/api/integrate/hadolint")
async def run_hadolint(dockerfile_path: str = "Dockerfile"):
    """
    Run hadolint on specified Dockerfile to check for best practices.
    """
    try:
        success, output = integrations.run_hadolint(dockerfile_path)
        return {
            "success": success,
            "output": output,
            "message": "Dockerfile linted successfully" if success else "hadolint found issues"
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/api/integrate/docker-compose-validate")
async def validate_docker_compose():
    """
    Validate docker-compose.yml configuration.
    """
    try:
        success, output = integrations.validate_docker_compose()
        return {
            "success": success,
            "output": output,
            "message": "docker-compose.yml is valid" if success else "docker-compose.yml has issues"
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/api/history")
async def get_fix_history(limit: int = 50):
    """
    Get history of fixes applied by Doctor.
    """
    try:
        history = doctor_logger.get_fix_history(limit)
        return {
            "total": len(history),
            "fixes": history
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/api/stats")
async def get_stats():
    """
    Get statistics about Doctor's performance.
    """
    try:
        stats = doctor_logger.get_stats()
        return stats
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
