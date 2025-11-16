import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { apiRequest } from '../utils/api';

// Projects API client wrapper
const projectsApi = {
  getAll: () => apiRequest('/api/logs/projects'),
  create: (data) => apiRequest('/api/logs/projects', {
    method: 'POST',
    body: JSON.stringify(data),
  }),
  regenerateKey: (id) => apiRequest(`/api/logs/projects/${id}/regenerate-key`, {
    method: 'POST',
  }),
  deactivate: (id) => apiRequest(`/api/logs/projects/${id}`, {
    method: 'DELETE',
  }),
};

export default function ProjectsPage() {
  const [projects, setProjects] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newApiKey, setNewApiKey] = useState(null);
  const [regeneratedKey, setRegeneratedKey] = useState(null);

  // Form state for creating new project
  const [formData, setFormData] = useState({
    name: '',
    slug: '',
    description: '',
    repository_url: ''
  });
  const [formErrors, setFormErrors] = useState({});

  useEffect(() => {
    fetchProjects();
  }, []);

  const fetchProjects = async () => {
    try {
      setLoading(true);
      const data = await projectsApi.getAll();
      setProjects(data || []);
      setError(null);
    } catch (err) {
      console.error('Failed to fetch projects:', err);
      setError(err.message || 'Failed to load projects');
    } finally {
      setLoading(false);
    }
  };

  const validateForm = () => {
    const errors = {};
    
    if (!formData.name.trim()) {
      errors.name = 'Project name is required';
    }
    
    if (!formData.slug.trim()) {
      errors.slug = 'Project slug is required';
    } else if (!/^[a-z0-9-]+$/.test(formData.slug)) {
      errors.slug = 'Slug must contain only lowercase letters, numbers, and hyphens';
    }
    
    if (formData.repository_url && !/^https?:\/\/.+/.test(formData.repository_url)) {
      errors.repository_url = 'Repository URL must start with http:// or https://';
    }
    
    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleCreateProject = async (e) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }
    
    try {
      const data = await projectsApi.create(formData);
      
      // Store the API key (shown only once)
      setNewApiKey(data.api_key);
      
      // Refresh projects list
      await fetchProjects();
      
      // Reset form
      setFormData({
        name: '',
        slug: '',
        description: '',
        repository_url: ''
      });
      setFormErrors({});
    } catch (err) {
      console.error('Failed to create project:', err);
      setError(err.message || 'Failed to create project');
    }
  };

  const handleRegenerateKey = async (projectId, projectName) => {
    if (!confirm(`Are you sure you want to regenerate the API key for "${projectName}"? The old key will stop working immediately.`)) {
      return;
    }
    
    try {
      const data = await projectsApi.regenerateKey(projectId);
      setRegeneratedKey({
        projectName,
        apiKey: data.api_key
      });
      
      // Refresh projects list
      await fetchProjects();
    } catch (err) {
      console.error('Failed to regenerate key:', err);
      setError(err.message || 'Failed to regenerate API key');
    }
  };

  const handleDeactivateProject = async (projectId, projectName) => {
    if (!confirm(`Are you sure you want to deactivate "${projectName}"? This will stop accepting logs for this project.`)) {
      return;
    }
    
    try {
      await projectsApi.deactivate(projectId);
      
      // Refresh projects list
      await fetchProjects();
    } catch (err) {
      console.error('Failed to deactivate project:', err);
      setError(err.message || 'Failed to deactivate project');
    }
  };

  const closeApiKeyModal = () => {
    setNewApiKey(null);
    setRegeneratedKey(null);
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    alert('API key copied to clipboard!');
  };

  return (
    <div className="container-fluid py-4">
      <div className="row mb-4">
        <div className="col">
          <h1 className="h2 mb-1">Cross-Repository Logging Projects</h1>
          <p className="text-muted">
            Manage API keys for external applications to send logs to DevSmith
          </p>
        </div>
        <div className="col-auto">
          <button
            className="btn btn-primary"
            onClick={() => setShowCreateModal(true)}
          >
            <i className="bi bi-plus-circle me-2"></i>
            Create Project
          </button>
        </div>
      </div>

      {error && (
        <div className="alert alert-danger alert-dismissible fade show" role="alert">
          <i className="bi bi-exclamation-triangle me-2"></i>
          {error}
          <button
            type="button"
            className="btn-close"
            onClick={() => setError(null)}
          ></button>
        </div>
      )}

      {loading ? (
        <div className="text-center py-5">
          <div className="spinner-border text-primary" role="status">
            <span className="visually-hidden">Loading...</span>
          </div>
        </div>
      ) : projects.length === 0 ? (
        <div className="card">
          <div className="card-body text-center py-5">
            <i className="bi bi-inbox display-1 text-muted"></i>
            <h3 className="mt-3">No Projects Yet</h3>
            <p className="text-muted mb-4">
              Create your first project to start sending logs from external applications
            </p>
            <button
              className="btn btn-primary"
              onClick={() => setShowCreateModal(true)}
            >
              <i className="bi bi-plus-circle me-2"></i>
              Create Your First Project
            </button>
          </div>
        </div>
      ) : (
        <div className="row g-4">
          {projects.map((project) => (
            <div key={project.id} className="col-12 col-md-6 col-xl-4">
              <div className="card h-100">
                <div className="card-body">
                  <div className="d-flex justify-content-between align-items-start mb-3">
                    <div>
                      <h5 className="card-title mb-1">{project.name}</h5>
                      <code className="text-muted small">{project.slug}</code>
                    </div>
                    <span
                      className={`badge ${
                        project.is_active ? 'bg-success' : 'bg-secondary'
                      }`}
                    >
                      {project.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </div>

                  {project.description && (
                    <p className="card-text text-muted small mb-3">
                      {project.description}
                    </p>
                  )}

                  {project.repository_url && (
                    <div className="mb-3">
                      <a
                        href={project.repository_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-decoration-none small"
                      >
                        <i className="bi bi-github me-1"></i>
                        Repository
                      </a>
                    </div>
                  )}

                  <div className="mb-3">
                    <small className="text-muted">
                      Created {new Date(project.created_at).toLocaleDateString()}
                    </small>
                  </div>

                  <div className="d-flex gap-2">
                    <button
                      className="btn btn-sm btn-outline-primary flex-grow-1"
                      onClick={() => window.location.href = `/logs?project_id=${project.id}`}
                    >
                      <i className="bi bi-list-ul me-1"></i>
                      View Logs
                    </button>
                    <button
                      className="btn btn-sm btn-outline-secondary"
                      onClick={() => handleRegenerateKey(project.id, project.name)}
                      title="Regenerate API Key"
                    >
                      <i className="bi bi-arrow-clockwise"></i>
                    </button>
                    {project.is_active && (
                      <button
                        className="btn btn-sm btn-outline-danger"
                        onClick={() => handleDeactivateProject(project.id, project.name)}
                        title="Deactivate Project"
                      >
                        <i className="bi bi-x-circle"></i>
                      </button>
                    )}
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create Project Modal */}
      {showCreateModal && (
        <div
          className="modal show d-block"
          tabIndex="-1"
          style={{ backgroundColor: 'rgba(0,0,0,0.5)' }}
        >
          <div className="modal-dialog modal-dialog-centered">
            <div className="modal-content">
              <div className="modal-header">
                <h5 className="modal-title">Create New Project</h5>
                <button
                  type="button"
                  className="btn-close"
                  onClick={() => {
                    setShowCreateModal(false);
                    setFormErrors({});
                  }}
                ></button>
              </div>
              <form onSubmit={handleCreateProject}>
                <div className="modal-body">
                  <div className="mb-3">
                    <label htmlFor="name" className="form-label">
                      Project Name <span className="text-danger">*</span>
                    </label>
                    <input
                      type="text"
                      className={`form-control ${formErrors.name ? 'is-invalid' : ''}`}
                      id="name"
                      value={formData.name}
                      onChange={(e) =>
                        setFormData({ ...formData, name: e.target.value })
                      }
                      placeholder="My Application"
                    />
                    {formErrors.name && (
                      <div className="invalid-feedback">{formErrors.name}</div>
                    )}
                  </div>

                  <div className="mb-3">
                    <label htmlFor="slug" className="form-label">
                      Project Slug <span className="text-danger">*</span>
                    </label>
                    <input
                      type="text"
                      className={`form-control ${formErrors.slug ? 'is-invalid' : ''}`}
                      id="slug"
                      value={formData.slug}
                      onChange={(e) =>
                        setFormData({ ...formData, slug: e.target.value })
                      }
                      placeholder="my-application"
                    />
                    <small className="form-text text-muted">
                      Used in API requests. Lowercase letters, numbers, and hyphens only.
                    </small>
                    {formErrors.slug && (
                      <div className="invalid-feedback">{formErrors.slug}</div>
                    )}
                  </div>

                  <div className="mb-3">
                    <label htmlFor="description" className="form-label">
                      Description
                    </label>
                    <textarea
                      className="form-control"
                      id="description"
                      rows="2"
                      value={formData.description}
                      onChange={(e) =>
                        setFormData({ ...formData, description: e.target.value })
                      }
                      placeholder="Brief description of this project..."
                    ></textarea>
                  </div>

                  <div className="mb-3">
                    <label htmlFor="repository_url" className="form-label">
                      Repository URL
                    </label>
                    <input
                      type="text"
                      className={`form-control ${
                        formErrors.repository_url ? 'is-invalid' : ''
                      }`}
                      id="repository_url"
                      value={formData.repository_url}
                      onChange={(e) =>
                        setFormData({ ...formData, repository_url: e.target.value })
                      }
                      placeholder="https://github.com/username/repo"
                    />
                    {formErrors.repository_url && (
                      <div className="invalid-feedback">
                        {formErrors.repository_url}
                      </div>
                    )}
                  </div>
                </div>
                <div className="modal-footer">
                  <button
                    type="button"
                    className="btn btn-secondary"
                    onClick={() => {
                      setShowCreateModal(false);
                      setFormErrors({});
                    }}
                  >
                    Cancel
                  </button>
                  <button type="submit" className="btn btn-primary">
                    <i className="bi bi-plus-circle me-2"></i>
                    Create Project
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* API Key Display Modal (shown once after creation or regeneration) */}
      {(newApiKey || regeneratedKey) && (
        <div
          className="modal show d-block"
          tabIndex="-1"
          style={{ backgroundColor: 'rgba(0,0,0,0.5)' }}
        >
          <div className="modal-dialog modal-dialog-centered">
            <div className="modal-content">
              <div className="modal-header bg-success text-white">
                <h5 className="modal-title">
                  <i className="bi bi-check-circle me-2"></i>
                  {newApiKey ? 'Project Created!' : 'API Key Regenerated'}
                </h5>
              </div>
              <div className="modal-body">
                <div className="alert alert-warning" role="alert">
                  <i className="bi bi-exclamation-triangle me-2"></i>
                  <strong>Important:</strong> Copy this API key now. You won't be able to
                  see it again!
                </div>

                {regeneratedKey && (
                  <p className="mb-3">
                    New API key for <strong>{regeneratedKey.projectName}</strong>:
                  </p>
                )}

                <div className="input-group mb-3">
                  <input
                    type="text"
                    className="form-control font-monospace"
                    value={newApiKey || regeneratedKey?.apiKey}
                    readOnly
                  />
                  <button
                    className="btn btn-outline-secondary"
                    type="button"
                    onClick={() =>
                      copyToClipboard(newApiKey || regeneratedKey?.apiKey)
                    }
                  >
                    <i className="bi bi-clipboard"></i>
                  </button>
                </div>

                <h6>Usage Example:</h6>
                <pre className="bg-light p-3 rounded">
                  <code>
{`curl -X POST http://your-devsmith-url/api/logs/batch \\
  -H "Authorization: Bearer ${newApiKey || regeneratedKey?.apiKey}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "project_slug": "${formData.slug || 'your-project-slug'}",
    "logs": [{
      "timestamp": "2025-11-11T16:40:00Z",
      "level": "info",
      "message": "Application started",
      "service_name": "api-server",
      "context": {"version": "1.0.0"}
    }]
  }'`}
                  </code>
                </pre>
              </div>
              <div className="modal-footer">
                <button
                  type="button"
                  className="btn btn-primary"
                  onClick={closeApiKeyModal}
                >
                  I've Copied the Key
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
