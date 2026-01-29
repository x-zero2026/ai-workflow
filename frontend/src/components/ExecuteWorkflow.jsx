import React, { useState } from 'react'
import api from '../api'
import './ExecuteWorkflow.css'

function ExecuteWorkflow({ workflow, onClose }) {
  // Build initial headers with Authorization
  const initialHeaders = {
    "Authorization": `Bearer ${workflow.bearer_token}`,
    "Content-Type": "application/json",
    ...(workflow.headers || {})
  }

  const [parameters, setParameters] = useState(
    JSON.stringify(workflow.parameters || {}, null, 2)
  )
  const [result, setResult] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  // Function to mask bearer token in request display
  const maskBearerToken = (obj) => {
    if (!obj) return obj
    
    const masked = JSON.parse(JSON.stringify(obj)) // Deep clone
    
    if (masked.headers && masked.headers.Authorization) {
      const authValue = masked.headers.Authorization
      if (authValue.startsWith('Bearer ')) {
        masked.headers.Authorization = 'Bearer ' + '*'.repeat(authValue.length - 7)
      }
    }
    
    return masked
  }

  const handleExecute = async () => {
    setError(null)
    setResult(null)

    // Validate JSON
    try {
      JSON.parse(parameters)
    } catch (e) {
      setError('Invalid JSON format in parameters')
      return
    }

    setLoading(true)

    try {
      const response = await api.executeWorkflow(workflow.workflow_id, {
        parameters: JSON.parse(parameters),
        headers: initialHeaders
      })

      if (response.success) {
        // Mask bearer token in the request before displaying
        const maskedResult = {
          ...response.data,
          request: maskBearerToken(response.data.request)
        }
        setResult(maskedResult)
      }
    } catch (err) {
      setError(err.error || 'Failed to execute workflow')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="modal-overlay">
      {/* Loading Overlay */}
      {loading && (
        <div className="loading-overlay">
          <div className="loading-spinner"></div>
        </div>
      )}

      <div className="modal-content execute-workflow">
        <div className="modal-header">
          <h2>Execute: {workflow.workflow_name}</h2>
          <button className="btn-close" onClick={onClose}>×</button>
        </div>

        <div className="execute-form">
          <div className="form-group">
            <label>Parameters (JSON)</label>
            <textarea
              value={parameters}
              onChange={(e) => setParameters(e.target.value)}
              rows="8"
              className="code-editor"
            />
          </div>

          <div className="form-actions">
            <button 
              className="btn btn-primary" 
              onClick={handleExecute}
              disabled={loading}
            >
              {loading ? 'Executing...' : 'Execute'}
            </button>
          </div>

          {error && (
            <div className="error-message">
              {error}
            </div>
          )}

          {result && (
            <div className="execute-result">
              <h3>Request</h3>
              <div className="result-section">
                <pre>{JSON.stringify(result.request, null, 2)}</pre>
              </div>

              <h3>Response</h3>
              <div className="result-section">
                <div className="response-status">
                  Status: <span className={result.response.status < 400 ? 'success' : 'error'}>
                    {result.response.status} {result.response.status_text}
                  </span>
                </div>
                <pre>{JSON.stringify(result.response.body, null, 2)}</pre>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default ExecuteWorkflow
