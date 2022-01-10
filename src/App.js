import './App.css';
import './util/responsive.css';
import MainLayout from './layouts/main';

function App() {
  return (
      <div className="App">
        <MainLayout extraRoutes={[]} extraNavigation={[]}/>
      </div>
  );
}

export default App;
