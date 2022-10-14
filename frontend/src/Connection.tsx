import {useEffect, useState} from "react";
import {CheckNetIsolation, ProxyingInfo, Terminate} from "../wailsjs/go/main/App";
import {main} from "../wailsjs/go/models";
import {useNavigate} from "react-router-dom";
import {BrowserOpenURL} from "../wailsjs/runtime";
import {LoopbackWarning} from "./Loopback";

function Connection() {
    const navigate = useNavigate()

    const [proxyingInfo, setProxyingInfo] = useState<main.ProxyInfo>({
        local_address: "", remote_address: "",
    })
    const [checkNetIsolation, setCheckNetIsolation] = useState(true)
    useEffect(() => {
        ProxyingInfo().then(result => setProxyingInfo(result))
        CheckNetIsolation().then(result => setCheckNetIsolation(result))
    }, [])

    return (
        <div>
            <h1 className={"text-slate-900 font-extrabold text-5xl tracking-tight dark:text-white"}>
                Tedac is up and running! ✨
            </h1>
            <div className="mt-12">
                <p className="text-md text-slate-600 max-w-xl dark:text-slate-400">
                    Tedac is now proxying your connection to
                    <code
                        className={"ml-1 text-slate-900 dark:text-blue-200 opacity-50 text-md"}>{proxyingInfo.remote_address}</code>.
                </p>
                <p className="mt-4 text-md text-slate-600 max-w-xl dark:text-slate-400">
                    In order to connect through Tedac, add and join through the relay server in your Minecraft client.
                    This will allow Tedac to intercept and translate your packets.
                </p>
                <div className={"mt-4"}>
                    <p className="text-md text-slate-600 max-w-xl dark:text-slate-400">
                        For support,
                        <a href="#" onClick={() => BrowserOpenURL("https://discord.gg/7Y4ruNgjgt")}
                           className="ml-1 mt-4 mr-1 text-md font-semibold text-sky-600 max-w-xl dark:text-sky-400">join
                            our Discord server</a>
                        and open a ticket.
                    </p>
                </div>
                <div className={"mt-8 flex flex-row"}>
                    <button
                        onClick={() => BrowserOpenURL(`minecraft://?addExternalServer=Tedac (${proxyingInfo.remote_address.split(":")[0]})|${proxyingInfo.local_address}`)}
                        className="text-white bg-slate-900 hover:bg-slate-800 focus:ring-4 focus:outline-none font-medium rounded-lg text-sm w-full sm:w-auto px-5 py-2.5 text-center dark:bg-green-600 dark:hover:bg-green-700 dark:focus:ring-green-800">
                        Add to Minecraft
                    </button>
                    <button
                        onClick={() => Terminate().then(() => navigate("/"))}
                        className="ml-3 text-white bg-slate-900 hover:bg-slate-800 focus:ring-4 focus:outline-none font-medium rounded-lg text-sm w-full sm:w-auto px-5 py-2.5 text-center dark:bg-red-500 dark:hover:bg-red-600 dark:focus:ring-red-700">
                        Disconnect
                    </button>
                </div>
                {!checkNetIsolation ?
                    <LoopbackWarning path={"/connection"} navigate={navigate}></LoopbackWarning> : <></>}
            </div>
        </div>
    )
}

export default Connection;
