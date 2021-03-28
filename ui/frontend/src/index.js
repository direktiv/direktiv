import React from 'react';
import ReactDOM from 'react-dom';
import './css/index.css';
import AppRouter from './components/app/app-router';
import reportWebVitals from './reportWebVitals';

ReactDOM.render(
    <AppRouter/>,
    document.getElementById('root')
);

reportWebVitals();
