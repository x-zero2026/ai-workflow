import axios from 'axios'

// AI Workflow API (AWS Lambda)
const workflowApi = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  timeout: 60000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// DID Login API (for projects)
const loginApi = axios.create({
  baseURL: import.meta.env.VITE_LOGIN_API_BASE_URL || 'http://localhost:8080',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor - add token to both APIs
const addTokenInterceptor = (config) => {
  const token = localStorage.getItem('xzero_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
}

// Response interceptor - handle errors
const handleErrorInterceptor = (error) => {
  if (error.response?.status === 401) {
    // Token expired, redirect to login
    localStorage.removeItem('xzero_token')
    window.location.href = '/'
  }
  return Promise.reject(error.response?.data || { error: error.message })
}

workflowApi.interceptors.request.use(addTokenInterceptor, (error) => Promise.reject(error))
workflowApi.interceptors.response.use((response) => response.data, handleErrorInterceptor)

loginApi.interceptors.request.use(addTokenInterceptor, (error) => Promise.reject(error))
loginApi.interceptors.response.use((response) => response.data, handleErrorInterceptor)

export default {
  // Project APIs (from DID Login)
  getProjects: () => loginApi.get('/api/projects'),

  // Workflow APIs (from AI Workflow backend)
  getWorkflows: (projectId) => workflowApi.get(`/api/projects/${projectId}/workflows`),
  
  createWorkflow: (data) => workflowApi.post('/api/workflows', data),
  
  updateWorkflow: (workflowId, data) => workflowApi.put(`/api/workflows/${workflowId}`, data),
  
  deleteWorkflow: (workflowId) => workflowApi.delete(`/api/workflows/${workflowId}`),
  
  executeWorkflow: (workflowId, data) => workflowApi.post(`/api/workflows/${workflowId}/execute`, data),
  
  shareWorkflow: (workflowId, isShared) => workflowApi.put(`/api/workflows/${workflowId}/share`, { is_shared: isShared }),
  
  hideWorkflow: (projectId, workflowId, isHidden) => 
    workflowApi.put(`/api/projects/${projectId}/workflows/${workflowId}/hide`, { is_hidden: isHidden }),
}
