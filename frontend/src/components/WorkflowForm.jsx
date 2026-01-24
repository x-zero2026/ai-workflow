import React, { useState } from 'react'
import './WorkflowForm.css'

function WorkflowForm({ projectId, workflow, onSubmit, onCancel }) {
  const isEdit = !!workflow

  const [formData, setFormData] = useState({
    workflow_name: workflow?.workflow_name || '',
    description: workflow?.description || '',
    source: workflow?.source || 'coze',
    template_name: workflow?.template_name || 'workflow',
    http_method: workflow?.http_method || 'POST',
    base_url: workflow?.base_url || '',
    bearer_token: workflow?.bearer_token || '',
    external_workflow_id: workflow?.external_workflow_id || '',
    parameters: workflow?.parameters ? JSON.stringify(workflow.parameters, null, 2) : '{}',
    headers: workflow?.headers ? JSON.stringify(workflow.headers, null, 2) : '{}',
    project_id: projectId
  })

  const [errors, setErrors] = useState({})
  const [loading, setLoading] = useState(false)

  const handleChange = (e) => {
    const { name, value } = e.target
    setFormData(prev => ({ ...prev, [name]: value }))
    // Clear error for this field
    if (errors[name]) {
      setErrors(prev => ({ ...prev, [name]: null }))
    }
  }

  const validate = () => {
    const newErrors = {}

    if (!formData.workflow_name.trim()) {
      newErrors.workflow_name = 'Workflow name is required'
    }
    if (!formData.description.trim()) {
      newErrors.description = 'Description is required'
    }
    if (!formData.base_url.trim()) {
      newErrors.base_url = 'Base URL is required'
    }
    if (!formData.bearer_token.trim()) {
      newErrors.bearer_token = 'Bearer token is required'
    }
    if (!formData.external_workflow_id.trim()) {
      newErrors.external_workflow_id = 'Workflow ID is required'
    }

    // Validate JSON
    try {
      JSON.parse(formData.parameters)
    } catch (e) {
      newErrors.parameters = 'Invalid JSON format'
    }

    try {
      JSON.parse(formData.headers)
    } catch (e) {
      newErrors.headers = 'Invalid JSON format'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e) => {
    e.preventDefault()

    if (!validate()) {
      return
    }

    setLoading(true)

    try {
      const submitData = {
        ...formData,
        parameters: JSON.parse(formData.parameters),
        headers: JSON.parse(formData.headers)
      }

      await onSubmit(submitData)
    } catch (err) {
      alert(err.error || 'Failed to save workflow')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="modal-overlay">
      <div className="modal-content workflow-form">
        <div className="modal-header">
          <h2>{isEdit ? 'Edit Workflow' : 'Create Workflow'}</h2>
          <button className="btn-close" onClick={onCancel}>×</button>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Workflow Name *</label>
            <input
              type="text"
              name="workflow_name"
              value={formData.workflow_name}
              onChange={handleChange}
              className={errors.workflow_name ? 'error' : ''}
            />
            {errors.workflow_name && <span className="error-message">{errors.workflow_name}</span>}
          </div>

          <div className="form-group">
            <label>Description *</label>
            <textarea
              name="description"
              value={formData.description}
              onChange={handleChange}
              rows="3"
              className={errors.description ? 'error' : ''}
            />
            {errors.description && <span className="error-message">{errors.description}</span>}
          </div>

          <div className="form-row">
            <div className="form-group">
              <label>Source *</label>
              <select name="source" value={formData.source} onChange={handleChange}>
                <option value="coze">Coze</option>
                <option value="n8n">n8n</option>
              </select>
            </div>

            <div className="form-group">
              <label>Template *</label>
              <select name="template_name" value={formData.template_name} onChange={handleChange}>
                <option value="workflow">Workflow</option>
                <option value="streamflow">Streamflow</option>
              </select>
            </div>

            <div className="form-group">
              <label>HTTP Method *</label>
              <select name="http_method" value={formData.http_method} onChange={handleChange}>
                <option value="GET">GET</option>
                <option value="POST">POST</option>
                <option value="PUT">PUT</option>
              </select>
            </div>
          </div>

          <div className="form-group">
            <label>Base URL *</label>
            <input
              type="text"
              name="base_url"
              value={formData.base_url}
              onChange={handleChange}
              placeholder="https://api.example.com/v1/workflow/execute"
              className={errors.base_url ? 'error' : ''}
            />
            {errors.base_url && <span className="error-message">{errors.base_url}</span>}
          </div>

          <div className="form-group">
            <label>Bearer Token *</label>
            <input
              type="password"
              name="bearer_token"
              value={formData.bearer_token}
              onChange={handleChange}
              className={errors.bearer_token ? 'error' : ''}
            />
            {errors.bearer_token && <span className="error-message">{errors.bearer_token}</span>}
          </div>

          <div className="form-group">
            <label>External Workflow ID *</label>
            <input
              type="text"
              name="external_workflow_id"
              value={formData.external_workflow_id}
              onChange={handleChange}
              placeholder="123456"
              className={errors.external_workflow_id ? 'error' : ''}
            />
            {errors.external_workflow_id && <span className="error-message">{errors.external_workflow_id}</span>}
          </div>

          <div className="form-group">
            <label>Parameters (JSON)</label>
            <textarea
              name="parameters"
              value={formData.parameters}
              onChange={handleChange}
              rows="5"
              className={errors.parameters ? 'error' : ''}
              placeholder='{"key": "value"}'
            />
            {errors.parameters && <span className="error-message">{errors.parameters}</span>}
          </div>

          <div className="form-group">
            <label>Headers (JSON)</label>
            <textarea
              name="headers"
              value={formData.headers}
              onChange={handleChange}
              rows="5"
              className={errors.headers ? 'error' : ''}
              placeholder='{"X-Custom-Header": "value"}'
            />
            {errors.headers && <span className="error-message">{errors.headers}</span>}
          </div>

          <div className="form-actions">
            <button type="button" className="btn btn-secondary" onClick={onCancel}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={loading}>
              {loading ? 'Saving...' : (isEdit ? 'Update' : 'Create')}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default WorkflowForm
