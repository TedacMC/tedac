import {useNavigate} from "react-router-dom";
import {BrowserOpenURL} from "../wailsjs/runtime";

function Error() {
    const navigate = useNavigate()
    const params = new URLSearchParams(window.location.search);

    const error = params.get("error") || "";
    const knownSolutions: {[index: string]: string } = {
        "no such host": "An invalid server address was provided. Check that you have entered the address correctly before trying again.",
        "error obtaining XBOX live token": "Failed to authenticate with XBOX live. Attempting to connect again will solve the issue.",
        "i/o timeout": "It appears that this server is offline. Check that you have entered the address correctly before trying again.",
        "getaddrinfow": "An invalid port has been provided. Check that you have entered the address correctly before trying again.",
        "invalid port": "An invalid port has been provided. Check that you have entered the address correctly before trying again."
    };

    const guessSolution = () => {
        const solutions = Object.keys(knownSolutions).filter(part => error.indexOf(part) >= 0);
        return solutions.length > 0 ? knownSolutions[solutions[0]] : "Unable to provide a solution for this error.";
    };

    return (
        <div>
            <h1 className={"text-slate-900 font-extrabold text-5xl tracking-tight dark:text-white"}>
                Tedac ran into an error! ❌
            </h1>
            <div className="mt-12">
                <p className="mt-4 text-md text-slate-600 max-w-xl dark:text-slate-400">
                    There was an error while attempting to connect to the server. The error and a potential solution is listed below.
                </p>
                <br/>
                <code
                    onClick={() => navigator.clipboard.writeText(error)}
                    className={"ml-1 text-slate-900 dark:text-blue-100 opacity-50 text-md cursor-pointer"}>
                    {error}
                </code>
                <p className="mt-4 text-md text-slate-600 max-w-xl dark:text-slate-400">
                    {guessSolution()}
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
                        onClick={() => navigate(`/?address=${params.get("address") || ""}&port=${params.get("port") || "19132"}`)}
                        className="text-white bg-slate-900 hover:bg-slate-800 focus:ring-4 focus:outline-none font-medium rounded-lg text-sm w-full sm:w-auto px-5 py-2.5 text-center dark:bg-red-500 dark:hover:bg-red-600 dark:focus:ring-red-700">
                        Try Again
                    </button>
                </div>
            </div>
        </div>
    )
}

export default Error;
