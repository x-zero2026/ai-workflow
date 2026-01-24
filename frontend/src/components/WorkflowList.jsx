import React from 'react'
import WorkflowCard from './WorkflowCard'
import './WorkflowList.css'

function WorkflowList({ workflows, onExecute, onEdit, onDelete, onShare, onHide, userRole }) {
  if (workflows.length === 0) {
    return (
      <div className="workflow-list-empty">
        <p>No workflows found. Create your first workflow!</p>
      </div>
    )
  }

  return (
    <div className="workflow-list">
      {workflows.map(workflow => (
        <WorkflowCard
          key={workflow.workflow_id}
          workflow={workflow}
          onExecute={onExecute}
          onEdit={onEdit}
          onDelete={onDelete}
          onShare={onShare}
          onHide={onHide}
          userRole={userRole}
        />
      ))}
    </div>
  )
}

export default WorkflowList
