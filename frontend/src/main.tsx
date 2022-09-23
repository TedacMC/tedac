import {BrowserRouter, Route, Routes,} from "react-router-dom";

import React from 'react'
import {render} from "react-dom";

import Home from './Home';
import Connection from "./Connection";

import './style.css'

render(
    <div className={"pt-12 justify-center px-12 text-slate-500 dark:text-slate-400 bg-white dark:bg-gray-900"}>
        <BrowserRouter>
            <Routes>
                <Route path="/" element={<Home/>}/>
                <Route path="/connection" element={<Connection/>}/>
            </Routes>
        </BrowserRouter>
    </div>,
    document.getElementById("root")
);
