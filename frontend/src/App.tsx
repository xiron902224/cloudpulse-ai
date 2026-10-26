import './App.css';
import { Dashboard } from './components/Dashboard';

function App() {
  return (
    <div className="app">
      <header className="app-header">
        <div className="header-content">
          <div className="logo">
            <svg
              width="32"
              height="32"
              viewBox="0 0 100 100"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <circle cx="50" cy="50" r="45" fill="#667eea" />
              <path d="M30 50 L50 30 L70 50 L50 70 Z" fill="white" />
            </svg>
            <h1>CloudPulse AI</h1>
          </div>
          <p className="tagline">Sistema de Monitoreo Cloud con IA</p>
        </div>
      </header>

      <main className="app-main">
        <Dashboard />
      </main>

      <footer className="app-footer">
        <p>© 2026 CloudPulse AI · Built with ❤️ by xiron9090</p>
      </footer>
    </div>
  );
}

export default App;