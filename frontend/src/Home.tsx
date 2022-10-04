import {CheckNetIsolation, Connect} from "../wailsjs/go/main/App";
import {Quit} from "../wailsjs/runtime";

import {useEffect, useState} from "react";
import {useNavigate} from "react-router-dom";

import zeqaLogo from '../logos/zeqa.jpg';
import vasarLogo from '../logos/vasar.jpg';
import {LoopbackWarning} from "./Loopback";

function Home() {
    const navigate = useNavigate()

    const [showServers, setShowServers] = useState(false)
    const [connectionButton, setConnectionButton] = useState(true);
    const [connectionLoader, setConnectionLoader] = useState("none");

    const [address, setAddress] = useState("");
    const [port, setPort] = useState("19132");

    const [checkNetIsolation, setCheckNetIsolation] = useState(true)
    useEffect(() => {
        CheckNetIsolation().then(result => setCheckNetIsolation(result))
    }, [])

    const servers: {
        name: string;
        address: string;
        logo: string;
    }[] = [
        {
            name: "Zeqa",
            address: "zeqa.net",
            logo: zeqaLogo,
        },
        {
            name: "Vasar",
            address: "vasar.land",
            logo: vasarLogo,
        }
    ];

    return (
        <div>
            <div className={"flex flex-row"}>
                <h1 className={"text-slate-900 font-extrabold max-w-sm text-5xl tracking-tight dark:text-white"}>
                    Welcome to Tedac. 👋
                </h1>
                <p className="ml-12 mt-4 text-lg text-slate-600 max-w-3xl dark:text-slate-400">
                    Tedac is a multi-version proxy that lets you join any Minecraft: Bedrock Edition server on v1.12.0,
                    no effort required.
                </p>
            </div>
            <div className="mt-10">
                <div className="grid gap-6 mb-8 md:grid-rows-2">
                    <div className={"max-w-md"}>
                        <label className="block mb-2 text-sm font-medium text-gray-900 dark:text-gray-300">
                            IP Address
                        </label>
                        <input type="text" id="ip" autoComplete={"off"} value={address}
                               onFocus={() => {
                                   if (!connectionButton) {
                                       // We're connecting, so don't allow the user to change the address.
                                       return
                                   }
                                   setShowServers(true)
                               }}
                               onBlur={(e) => {
                                   if (e.relatedTarget === null) {
                                       setShowServers(false);
                                   }
                               }}
                               onChange={(e) => {
                                   if (!connectionButton) {
                                       // We're connecting, so don't allow the user to change the address.
                                       return
                                   }
                                   setAddress(e.target.value);
                                   setShowServers(false);
                               }}
                               className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                               placeholder="zeqa.net" required></input>
                        {showServers ?
                            <div
                                tabIndex={0}
                                className="grid grid-cols-1 grid-rows-2 grid-flow-col gap-x-2 mt-2 bg-slate-700 shadow-lg rounded-lg absolute">
                                {servers.map((server) => (
                                    <button
                                        className="pl-2.5 py-2.5 pr-32 flex flew-row items-center bg-transparent rounded-lg hover:bg-slate-800"
                                        onClick={() => {
                                            setAddress(server.address);
                                            setShowServers(false);
                                        }}>
                                        <img className="rounded-lg max-h-16 max-w-16" src={server.logo}/>
                                        <div className="ml-3 text-left">
                                            <p className="font-extrabold text-md">{server.name}</p>
                                            <p className="text-sm">{server.address}</p>
                                        </div>
                                    </button>
                                ))}
                            </div>
                            : null}
                    </div>
                    <div className={"max-w-xs"}>
                        <label className="block mb-2 text-sm font-medium text-gray-900 dark:text-gray-300">
                            Port
                        </label>
                        <input type="text" id="port" autoComplete={"off"} value={port}
                               onChange={(e) => {
                                   if (!connectionButton) {
                                       // We're connecting, so don't allow the user to change the port.
                                       return
                                   }
                                   setPort(e.target.value)
                               }}
                               className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                               required></input>
                    </div>
                    <div className={"flex flex-row"}>
                        <button
                            onClick={() => {
                                if (!address) {
                                    // Address not set, don't bother.
                                    return;
                                }

                                // Disable the button and show the loader.
                                setConnectionButton(false);
                                setConnectionLoader("inline");

                                // Connect through the backend.
                                Connect(address + ":" + port).then(() => {
                                    navigate("/connection");
                                }).catch((e) => {
                                    navigate("/error?error=" + e);
                                });
                            }} disabled={!connectionButton}
                            className="inline-flex items-center text-white bg-slate-900 hover:bg-slate-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm w-full sm:w-auto px-5 py-2.5 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800">
                            <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white"
                                 xmlns="http://www.w3.org/2000/svg"
                                 fill="none" viewBox="0 0 24 24" style={{display: connectionLoader}}>
                                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor"
                                        strokeWidth="4"/>
                                <path className="opacity-75" fill="currentColor"
                                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
                            </svg>
                            Connect through Tedac
                        </button>
                        <button
                            onClick={Quit}
                            className="ml-3 text-white bg-slate-900 hover:bg-slate-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm w-full sm:w-auto px-5 py-2.5 text-center dark:bg-red-500 dark:hover:bg-red-600 dark:focus:ring-red-700">
                            Exit
                        </button>
                    </div>
                </div>
            </div>
            {!checkNetIsolation ?
                <LoopbackWarning path={"/"} navigate={navigate}></LoopbackWarning> : <></>}
        </div>

    )
}

export default Home;
