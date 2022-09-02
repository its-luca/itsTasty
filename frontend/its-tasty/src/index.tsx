import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import reportWebVitals from './reportWebVitals';
import {BrowserRouter, Route, Routes} from "react-router-dom";
import {DummyPublic,DummyPrivate} from "./routes/dummyRoutes";
import {PrivateRoutes} from "./PrivateRoutes"
import {LoginPage} from "./routes/login";
import {RateDishByID} from "./routes/rateDishByID";


const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);
root.render(
  <React.StrictMode>
      <BrowserRouter>
          <link
              rel="stylesheet"
              href="https://cdn.jsdelivr.net/npm/bootstrap@5.2.0-beta1/dist/css/bootstrap.min.css"
              integrity="sha384-0evHe/X+R7YkIZDRvuzKMRqM+OrBnVFBL6DOitfPri4tjfHxaWutUpFmBp4vmVor"
              crossOrigin="anonymous"
          />
          <Routes>
              <Route path={"/login"} element={<LoginPage/>}/>
              <Route path={"/"} element={<App />}>

              <Route element={<PrivateRoutes/>}>
                      <Route path={"/dish/:id"} element={<RateDishByID/>}/>
                  </Route>
                  <Route path="public" element={<DummyPublic />} />
              </Route>
          </Routes>
      </BrowserRouter>
  </React.StrictMode>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
