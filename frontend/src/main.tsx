import {BrowserRouter, Route, Routes,} from "react-router-dom";

import React from 'react'
import {render} from "react-dom";

import Home from './Home';
import Connection from "./Connection";

import './style.css'
import Loopback from "./Loopback";

render(
    <div className={"flex flex-col justify-center px-12 text-slate-500 dark:text-slate-400 bg-white dark:bg-gray-900"}>
        <BrowserRouter>
            <Routes>
                <Route path="/" element={<Home/>}/>
                <Route path="/connection" element={<Connection/>}/>
                <Route path="/loopback" element={<Loopback/>}/>
            </Routes>
        </BrowserRouter>
    </div>,
    document.getElementById("root")
);
