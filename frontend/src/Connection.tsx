import {useEffect, useState} from "react";
import {ProxyingInfo, Terminate} from "../wailsjs/go/main/App";
import {main} from "../wailsjs/go/models";
import {useNavigate} from "react-router-dom";
import {BrowserOpenURL} from "../wailsjs/runtime";

function Connection() {
    const navigate = useNavigate()

    const [proxyingInfo, setProxyingInfo] = useState<main.ProxyInfo>({
        local_address: "", remote_address: "",
    })
    useEffect(() => {
        ProxyingInfo()
            .then(result => setProxyingInfo(result))
            .catch(() => navigate("/"))
    })

    return (
        <div>
            <div className={"flex flex-row"}>
                <h1 className={"text-slate-900 font-extrabold text-5xl tracking-tight dark:text-white"}>
                    Tedac is up and running! ✨
                </h1>
                <div className={"ml-48 mr-4 mt-5 flex flex-col"}>
                    <code className={"text-slate-800 dark:text-blue-100 opacity-50 text-2xl"}>connect to</code>
                    <code className={"text-slate-900 dark:text-blue-200 opacity-50 text-3xl"}>{proxyingInfo.local_address}</code>
                </div>
            </div>
            <div className="mt-12">
                <p className="text-md text-slate-600 max-w-xl dark:text-slate-400">
                    Tedac is now proxying your connection to
                    <code className={"ml-1 text-slate-900 dark:text-blue-200 opacity-50 text-md"}>{proxyingInfo.remote_address}</code>.
                </p>
                <p className="mt-4 text-md text-slate-600 max-w-xl dark:text-slate-400">
                    In order to connect to the server, join the proxy address above. Doing so will connect you through
                    Tedac instead of directly to the server, allowing Tedac to translate the packets from the latest
                    version to work with v1.12.0.
                </p>
                <div className={"mt-4"}>
                    <p className="text-md text-slate-600 max-w-xl dark:text-slate-400">
                        For support,
                        <a href="#" onClick={() => BrowserOpenURL("https://discord.gg/7Y4ruNgjgt")} className="ml-1 mt-4 mr-1 text-md font-semibold text-sky-600 max-w-xl dark:text-sky-400">join our Discord server</a>
                        and open a ticket.
                    </p>
                </div>
                <div className={"mt-8 flex flex-row"}>
                    <button
                        onClick={() => BrowserOpenURL(`minecraft://?addExternalServer=Tedac Relay|${proxyingInfo.local_address}`)}
                        className="text-white bg-slate-900 hover:bg-slate-800 focus:ring-4 focus:outline-none font-medium rounded-lg text-sm w-full sm:w-auto px-5 py-2.5 text-center dark:bg-green-600 dark:hover:bg-green-700 dark:focus:ring-green-800">
                        Add to Minecraft
                    </button>
                    <button
                        onClick={() => Terminate().then(() => navigate("/"))}
                        className="ml-3 text-white bg-slate-900 hover:bg-slate-800 focus:ring-4 focus:outline-none font-medium rounded-lg text-sm w-full sm:w-auto px-5 py-2.5 text-center dark:bg-red-500 dark:hover:bg-red-600 dark:focus:ring-red-700">
                        Disconnect
                    </button>
                </div>
            </div>
        </div>
    )
}

export default Connection;
