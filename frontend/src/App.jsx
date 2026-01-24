import React, { useState, useEffect } from 'react'
import api from './api'
import WorkflowList from './components/WorkflowList'
import WorkflowForm from './components/WorkflowForm'
import ExecuteWorkflow from './components/ExecuteWorkflow'
import './App.css'

function App() {
  const [projects, setProjects] = useState([])
  const [selectedProject, setSelectedProject] = useState(null)
  const [workflows, setWorkflows] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [selectedWorkflow, setSelectedWorkflow] = useState(null)
  const [userRole, setUserRole] = useState('member')

  useEffect(() => {
    // Check for token in URL (passed from DID Login)
    const urlParams = new URLSearchParams(window.location.search)
    const tokenFromUrl = urlParams.get('token')
    
    if (tokenFromUrl) {
      console.log('✅ Token received from URL:', tokenFromUrl.substring(0, 20) + '...')
      // Save token to localStorage with key 'xzero_token'
      localStorage.setItem('xzero_token', tokenFromUrl)
      console.log('✅ Token saved to localStorage as xzero_token')
      // Remove token from URL for security
      window.history.replaceState({}, document.title, window.location.pathname)
      console.log('✅ Token removed from URL')
      // Reload to apply the token
      window.location.reload()
      return
    }
    
    // Check if token exists in localStorage (key is 'xzero_token')
    const token = localStorage.getItem('xzero_token')
    console.log('Checking localStorage xzero_token:', token ? '✅ Found' : '❌ Not found')
    
    if (!token) {
      console.log('❌ No token found, redirecting to login')
      setError('请先登录 DID Login 系统。3秒后将跳转到登录页面...')
      setTimeout(() => {
        window.location.href = import.meta.env.VITE_DID_LOGIN_URL || 'http://localhost:3000/dashboard'
      }, 3000)
      return
    }
    
    console.log('✅ Token found, loading projects')
    loadProjects()
  }, [])

  useEffect(() => {
    if (selectedProject) {
      loadWorkflows()
      checkUserRole()
    }
  }, [selectedProject])

  const loadProjects = async () => {
    try {
      setLoading(true)
      const response = await api.getProjects()
      if (response.success) {
        setProjects(response.data)
        if (response.data.length > 0) {
          setSelectedProject(response.data[0])
        }
      }
    } catch (err) {
      setError(err.error || 'Failed to load projects')
    } finally {
      setLoading(false)
    }
  }

  const loadWorkflows = async () => {
    if (!selectedProject) return
    
    try {
      setLoading(true)
      const response = await api.getWorkflows(selectedProject.project_id)
      if (response.success) {
        setWorkflows(response.data)
      }
    } catch (err) {
      setError(err.error || 'Failed to load workflows')
    } finally {
      setLoading(false)
    }
  }

  const checkUserRole = () => {
    if (selectedProject) {
      setUserRole(selectedProject.role)
    }
  }

  const handleCreateWorkflow = async (workflowData) => {
    try {
      let response
      if (selectedWorkflow) {
        // Update existing workflow
        response = await api.updateWorkflow(selectedWorkflow.workflow_id, workflowData)
      } else {
        // Create new workflow
        response = await api.createWorkflow(workflowData)
      }
      
      if (response.success || response.workflow_id) {
        setShowCreateForm(false)
        setSelectedWorkflow(null)
        loadWorkflows()
      }
    } catch (err) {
      throw err
    }
  }

  const handleExecuteWorkflow = (workflow) => {
    setSelectedWorkflow(workflow)
  }

  const handleCloseExecute = () => {
    setSelectedWorkflow(null)
  }

  const handleEditWorkflow = (workflow) => {
    setSelectedWorkflow(workflow)
    setShowCreateForm(true)
  }

  const handleDeleteWorkflow = async (workflowId) => {
    try {
      await api.deleteWorkflow(workflowId)
      loadWorkflows()
    } catch (err) {
      setError(err.error || 'Failed to delete workflow')
    }
  }

  const handleShareWorkflow = async (workflowId, isShared) => {
    try {
      await api.shareWorkflow(workflowId, isShared)
      loadWorkflows()
    } catch (err) {
      setError(err.error || 'Failed to share workflow')
    }
  }

  const handleHideWorkflow = async (workflowId, isHidden) => {
    try {
      await api.hideWorkflow(selectedProject.project_id, workflowId, isHidden)
      loadWorkflows()
    } catch (err) {
      setError(err.error || 'Failed to hide workflow')
    }
  }

  return (
    <div className="app">
      <header className="app-header">
        <h1>AI Workflow Center</h1>
        <div className="project-selector">
          <label>Project: </label>
          <select 
            value={selectedProject?.project_id || ''} 
            onChange={(e) => {
              const project = projects.find(p => p.project_id === e.target.value)
              setSelectedProject(project)
            }}
          >
            {projects.map(project => (
              <option key={project.project_id} value={project.project_id}>
                {project.project_name} ({project.role})
              </option>
            ))}
          </select>
        </div>
      </header>

      <main className="app-main">
        {error && (
          <div className="error-message">
            {error}
            <button onClick={() => setError(null)}>×</button>
          </div>
        )}

        {userRole === 'admin' && (
          <div className="actions">
            <button 
              className="btn-primary" 
              onClick={() => setShowCreateForm(true)}
            >
              + Create Workflow
            </button>
          </div>
        )}

        {loading ? (
          <div className="loading">Loading...</div>
        ) : (
          <WorkflowList 
            workflows={workflows}
            onExecute={handleExecuteWorkflow}
            onEdit={handleEditWorkflow}
            onDelete={handleDeleteWorkflow}
            onShare={handleShareWorkflow}
            onHide={handleHideWorkflow}
            userRole={userRole}
          />
        )}
      </main>

      {showCreateForm && !selectedWorkflow && (
        <WorkflowForm
          projectId={selectedProject?.project_id}
          onSubmit={handleCreateWorkflow}
          onCancel={() => setShowCreateForm(false)}
        />
      )}

      {showCreateForm && selectedWorkflow && (
        <WorkflowForm
          projectId={selectedProject?.project_id}
          workflow={selectedWorkflow}
          onSubmit={handleCreateWorkflow}
          onCancel={() => {
            setShowCreateForm(false)
            setSelectedWorkflow(null)
          }}
        />
      )}

      {selectedWorkflow && !showCreateForm && (
        <ExecuteWorkflow
          workflow={selectedWorkflow}
          onClose={handleCloseExecute}
        />
      )}
    </div>
  )
}

export default App
