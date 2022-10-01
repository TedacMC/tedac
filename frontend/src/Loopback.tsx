import {useEffect, useState} from "react";
import {CheckNetIsolation, Terminate} from "../wailsjs/go/main/App";
import {useNavigate} from "react-router-dom";
import {BrowserOpenURL} from "../wailsjs/runtime";

function Loopback() {
    const navigate = useNavigate()

    const loopbackCommand = `CheckNetIsolation LoopbackExempt -a -n="Microsoft.MinecraftUWP_8wekyb3d8bbwe"`;
    const [checkNetIsolation, setCheckNetIsolation] = useState(true)
    useEffect(() => {
        CheckNetIsolation().then(result => setCheckNetIsolation(result))
    }, [])

    if (!checkNetIsolation) {
        return (
            <div>
                <h1 className={"text-slate-900 font-extrabold text-5xl tracking-tight dark:text-white"}>
                    Tedac is ready to be used! ✨
                </h1>
                <div className="mt-12">
                    <p className="text-md text-slate-600 max-w-xl dark:text-slate-400">
                        You are ready to connect to Tedac on this device!
                    </p>
                    <p className="mt-4 text-md text-slate-600 max-w-xl dark:text-slate-400">
                        This command is only required to be ran once, so you will not have to do this any more in the future. If you are still having issues connecting,
                        <a href="#"
                           onClick={() => BrowserOpenURL("https://discord.gg/7Y4ruNgjgt")}
                           className="ml-1 mt-4 mr-1 text-md font-semibold text-sky-600 max-w-xl dark:text-sky-400">
                            join our Discord server
                        </a>
                        and open a ticket.
                    </p>
                    <button
                        onClick={() => navigate(new URLSearchParams(window.location.search).get("path") || "/")}
                        className="mt-8 text-white bg-slate-900 hover:bg-slate-800 focus:ring-4 focus:outline-none font-medium rounded-lg text-sm w-full sm:w-auto px-5 py-2.5 text-center dark:bg-red-500 dark:hover:bg-red-600 dark:focus:ring-red-700">
                        Go Back
                    </button>
                </div>
            </div>
        )
    }

    return (
        <div>
            <h1 className={"text-slate-900 font-extrabold text-5xl tracking-tight dark:text-white"}>
                Enable Loopback Exemption 🌐
            </h1>
            <div className="mt-12">
                <p className="mt-4 text-md text-slate-600 max-w-xl dark:text-slate-400">
                    In order to connect to Tedac on this device, you must enable network isolation for Minecraft. Open an admin-level PowerShell session and run this command.
                </p>
                <br />
                <code
                    onClick={() => navigator.clipboard.writeText(loopbackCommand)}
                      className={"ml-1 text-slate-900 dark:text-blue-100 opacity-50 text-md cursor-pointer"}>
                    {loopbackCommand}
                </code>
                <p className="mt-4 text-md text-slate-600 max-w-xl dark:text-slate-400">
                    After running this command, click "Refresh" to see if you are ready to connect.
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
                        onClick={() => CheckNetIsolation().then(result => setCheckNetIsolation(result))}
                        className="text-white bg-slate-900 hover:bg-slate-800 focus:ring-4 focus:outline-none font-medium rounded-lg text-sm w-full sm:w-auto px-5 py-2.5 text-center dark:bg-green-600 dark:hover:bg-green-700 dark:focus:ring-green-800">
                        Refresh
                    </button>
                    <button
                        onClick={() => navigate("/")}
                        className="ml-3 text-white bg-slate-900 hover:bg-slate-800 focus:ring-4 focus:outline-none font-medium rounded-lg text-sm w-full sm:w-auto px-5 py-2.5 text-center dark:bg-red-500 dark:hover:bg-red-600 dark:focus:ring-red-700">
                        Go Back
                    </button>
                </div>
            </div>
        </div>
    )
}

export default Loopback;
