import React from 'react'
import WorkflowCard from './WorkflowCard'
import './WorkflowList.css'

function WorkflowList({ workflows, onExecute, onEdit, onDelete, onShare, onHide, userRole }) {
  if (workflows.length === 0) {
    return (
      <div className="workflow-list-empty">
        <p>暂无工作流。创建你的第一个工作流吧！</p>
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
