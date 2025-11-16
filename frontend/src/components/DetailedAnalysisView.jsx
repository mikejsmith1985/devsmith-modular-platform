import React from 'react';

/**
 * DetailedAnalysisView - Renders line-by-line code analysis with visual separation
 * - Code displayed in blue/gray
 * - AI explanations in chat bubbles (indigo/purple gradient)
 */
export default function DetailedAnalysisView({ analysis, fontSize = 'medium' }) {
  const fontSizes = {
    xsmall: '0.875rem',
    small: '1.0rem',
    medium: '1.125rem',
    large: '1.25rem',
    xlarge: '1.375rem'
  };

  // Parse the analysis text to extract line-by-line content
  // Expected format: alternating code lines and explanation blocks
  const parseAnalysis = (text) => {
    if (!text) return [];
    
    const lines = text.split('\n');
    const parsed = [];
    let currentBlock = { type: 'code', content: [], lineNumbers: [] };
    let lineNumber = 1;
    
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      
      // Detect explanation markers (AI responses typically start with markers)
      const isExplanation = line.trim().startsWith('//') || 
                           line.trim().startsWith('/*') ||
                           line.includes('Explanation:') ||
                           line.includes('Analysis:') ||
                           line.includes('Note:');
      
      if (isExplanation) {
        // Save current code block if it has content
        if (currentBlock.type === 'code' && currentBlock.content.length > 0) {
          parsed.push({ ...currentBlock });
          currentBlock = { type: 'explanation', content: [] };
        }
        
        // Add to explanation block
        if (currentBlock.type === 'explanation') {
          currentBlock.content.push(line.replace(/^\/\/|^\/\*|\*\/$/g, '').trim());
        } else {
          // Start new explanation block
          parsed.push({ ...currentBlock });
          currentBlock = { 
            type: 'explanation', 
            content: [line.replace(/^\/\/|^\/\*|\*\/$/g, '').trim()] 
          };
        }
      } else {
        // Save current explanation block if it has content
        if (currentBlock.type === 'explanation' && currentBlock.content.length > 0) {
          parsed.push({ ...currentBlock });
          currentBlock = { type: 'code', content: [], lineNumbers: [] };
        }
        
        // Add to code block
        if (line.trim()) {
          currentBlock.content.push(line);
          currentBlock.lineNumbers.push(lineNumber);
          lineNumber++;
        }
      }
    }
    
    // Add final block
    if (currentBlock.content.length > 0) {
      parsed.push(currentBlock);
    }
    
    return parsed;
  };

  const blocks = parseAnalysis(analysis);

  return (
    <div className="detailed-analysis-view" style={{ fontSize: fontSizes[fontSize] }}>
      {blocks.map((block, index) => (
        <div key={index} className="mb-3">
          {block.type === 'code' ? (
            // Code block - monospace font with line numbers, blue/gray tint
            <div 
              className="code-block p-3 rounded"
              style={{
                backgroundColor: 'rgba(59, 130, 246, 0.05)',
                border: '2px solid rgba(59, 130, 246, 0.15)',
                fontFamily: 'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, "Liberation Mono", monospace',
                color: '#1e40af',
                position: 'relative'
              }}
            >
              {block.content.map((line, lineIndex) => (
                <div 
                  key={lineIndex} 
                  className="code-line"
                  style={{
                    display: 'flex',
                    gap: '1rem',
                    paddingLeft: '0.5rem'
                  }}
                >
                  <span 
                    className="line-number" 
                    style={{
                      color: 'rgba(30, 64, 175, 0.4)',
                      minWidth: '3rem',
                      textAlign: 'right',
                      userSelect: 'none'
                    }}
                  >
                    {block.lineNumbers[lineIndex]}
                  </span>
                  <span style={{ flex: 1, whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
                    {line}
                  </span>
                </div>
              ))}
            </div>
          ) : (
            // Explanation block - chat bubble style with gradient
            <div 
              className="explanation-bubble p-3 rounded-3 position-relative"
              style={{
                background: 'linear-gradient(135deg, rgba(99, 102, 241, 0.1), rgba(139, 92, 246, 0.1))',
                border: '2px solid rgba(99, 102, 241, 0.2)',
                marginLeft: '3rem',
                color: '#4c1d95',
                lineHeight: '1.6'
              }}
            >
              {/* Chat bubble arrow */}
              <div
                style={{
                  position: 'absolute',
                  left: '-12px',
                  top: '20px',
                  width: 0,
                  height: 0,
                  borderTop: '10px solid transparent',
                  borderBottom: '10px solid transparent',
                  borderRight: '12px solid rgba(99, 102, 241, 0.2)'
                }}
              />
              <div
                style={{
                  position: 'absolute',
                  left: '-9px',
                  top: '21px',
                  width: 0,
                  height: 0,
                  borderTop: '9px solid transparent',
                  borderBottom: '9px solid transparent',
                  borderRight: '10px solid rgba(99, 102, 241, 0.1)'
                }}
              />
              
              {/* AI avatar icon */}
              <div 
                className="d-flex align-items-start gap-2"
                style={{ marginBottom: '0.5rem' }}
              >
                <span style={{ 
                  fontSize: '1.2em',
                  opacity: 0.8
                }}>
                  🤖
                </span>
                <div style={{ flex: 1 }}>
                  {block.content.map((line, lineIndex) => (
                    <p key={lineIndex} className="mb-2" style={{ wordWrap: 'break-word' }}>
                      {line}
                    </p>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>
      ))}
      
      {blocks.length === 0 && (
        <div className="text-center text-muted py-4">
          <p>No detailed analysis available yet.</p>
          <small>Run the analysis to see line-by-line explanations.</small>
        </div>
      )}
    </div>
  );
}
