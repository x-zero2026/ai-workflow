import React from 'react'
import './WorkflowCard.css'

function WorkflowCard({ workflow, onExecute, onEdit, onDelete, onShare, onHide, userRole }) {
  const isAdmin = userRole === 'admin'

  const handleExecute = () => {
    onExecute(workflow)
  }

  const handleEdit = () => {
    onEdit(workflow)
  }

  const handleDelete = () => {
    if (window.confirm(`Are you sure you want to delete "${workflow.workflow_name}"?`)) {
      onDelete(workflow.workflow_id)
    }
  }

  const handleShare = () => {
    const action = workflow.is_shared ? 'unshare' : 'share'
    if (window.confirm(`Are you sure you want to ${action} this workflow?`)) {
      onShare(workflow.workflow_id, !workflow.is_shared)
    }
  }

  const handleHide = () => {
    if (window.confirm(`Are you sure you want to hide "${workflow.workflow_name}"?`)) {
      onHide(workflow.workflow_id, true)
    }
  }

  return (
    <div className="workflow-card">
      <div className="workflow-card-header">
        <h3>{workflow.workflow_name}</h3>
        <div className="workflow-badges">
          <span className={`badge badge-${workflow.source}`}>{workflow.source}</span>
          <span className="badge badge-template">{workflow.template_name}</span>
          {workflow.is_shared && <span className="badge badge-shared">Shared</span>}
        </div>
      </div>

      <div className="workflow-card-body">
        <p className="workflow-description">{workflow.description}</p>
        
        <div className="workflow-meta">
          <div className="workflow-meta-item">
            <span className="label">Method:</span>
            <span className="value">{workflow.http_method}</span>
          </div>
          <div className="workflow-meta-item">
            <span className="label">URL:</span>
            <span className="value workflow-url" title={workflow.base_url}>
              {workflow.base_url}
            </span>
          </div>
        </div>
      </div>

      <div className="workflow-card-footer">
        <button className="btn btn-primary" onClick={handleExecute}>
          Execute
        </button>

        {isAdmin && (
          <div className="workflow-actions">
            <button className="btn btn-secondary" onClick={handleEdit}>
              Edit
            </button>
            <button className="btn btn-secondary" onClick={handleShare}>
              {workflow.is_shared ? 'Unshare' : 'Share'}
            </button>
            <button className="btn btn-secondary" onClick={handleHide}>
              Hide
            </button>
            <button className="btn btn-danger" onClick={handleDelete}>
              Delete
            </button>
          </div>
        )}
      </div>
    </div>
  )
}

export default WorkflowCard
