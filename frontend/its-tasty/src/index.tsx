import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import reportWebVitals from './reportWebVitals';
import {BrowserRouter, Route, Routes} from "react-router-dom";
import {PrivateRoutes} from "./PrivateRoutes"
import {LoginPage} from "./routes/login";
import {MergedDishViewByIDAdapter, MergeDishByIDAdapter, RateDishByID} from "./routes/rateDishByID";
import '@fontsource/roboto/300.css';
import '@fontsource/roboto/400.css';
import '@fontsource/roboto/500.css';
import '@fontsource/roboto/700.css';
import {ViewDishesAtDateURLAdapter} from "./routes/viewDishesAtDateURLAdapter";
import {ThemeSelector} from "./themeSelector";
import {DishGridView} from "./dishes/dishGridView";


const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);
root.render(
  <React.StrictMode>
      <BrowserRouter basename={new URL(process.env.REACT_APP_PUBLIC_URL!).pathname}>
          <ThemeSelector>
              <Routes>
                  <Route path={"/login"} element={<LoginPage/>}/>
                  <Route element={<PrivateRoutes/>}>
                      <Route path={"/"} element={<App />}>
                          <Route path={"/welcome"} element={<ViewDishesAtDateURLAdapter/>}/>
                          <Route path={"/dish/:id"} element={<RateDishByID/>}/>
                          <Route path={"dishesByDate/:dateString"} element={<ViewDishesAtDateURLAdapter/>}/>
                          <Route path={"dishesByDate/"} element={<ViewDishesAtDateURLAdapter/>}/>
                          <Route path={"/mergeCandidates/:id"} element={<MergeDishByIDAdapter/>}/>
                          <Route path={"/mergedDish/:id"} element={<MergedDishViewByIDAdapter/>}/>
                          <Route path={"/dishes"} element={<DishGridView/>}/>

                      </Route>
                  </Route>
              </Routes>
          </ThemeSelector>
      </BrowserRouter>
  </React.StrictMode>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
