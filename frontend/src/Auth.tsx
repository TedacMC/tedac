import {useEffect, useState} from "react";
import {useNavigate} from "react-router-dom";
import {FinishAuth, StartAuth} from "../wailsjs/go/main/App";
import {BrowserOpenURL} from "../wailsjs/runtime";

function Auth() {
    const navigate = useNavigate()
    const params = new URLSearchParams(window.location.search);

    const [authURL, setAuthURL] = useState("");
    const [finishing, setFinishing] = useState(false);
    const [error, setError] = useState("");

    useEffect(() => {
        StartAuth().then(url => {
            setAuthURL(url);
        }).catch(e => {
            setError(String(e));
        });
    }, []);

    const handleFinish = () => {
        setFinishing(true);
        setError("");
        FinishAuth().then(() => {
            const address = params.get("address") || "";
            const port = params.get("port") || "19132";
            const localPort = params.get("localPort") || "19132";
            navigate(`/?address=${address}&port=${port}&localPort=${localPort}`);
        }).catch(e => {
            setFinishing(false);
            setError(String(e));
        });
    };

    return (
        <div>
            <h1 className={"text-slate-900 font-extrabold text-5xl tracking-tight dark:text-white"}>
                Authentication Required
            </h1>
            <div className="mt-12">
                <p className="text-md text-slate-600 max-w-xl dark:text-slate-400">
                    Sign in with your Microsoft account to continue. A browser window has been opened for you.
                </p>
                {authURL ? (
                    <div className="mt-4">
                        <p className="text-md text-slate-600 max-w-xl dark:text-slate-400">
                            If the browser didn't open, click the button below or visit this URL:
                        </p>
                        <code
                            onClick={() => navigator.clipboard.writeText(authURL)}
                            className="mt-2 block ml-1 text-slate-900 dark:text-blue-100 opacity-50 text-sm cursor-pointer break-all">
                            {authURL}
                        </code>
                    </div>
                ) : null}
                {error ? (
                    <p className="mt-4 text-md text-red-400">{error}</p>
                ) : null}
                <div className={"mt-8 flex flex-row"}>
                    {authURL ? (
                        <button
                            onClick={() => BrowserOpenURL(authURL)}
                            className="text-white bg-slate-900 hover:bg-slate-800 focus:ring-4 focus:outline-none font-medium rounded-lg text-sm w-full sm:w-auto px-5 py-2.5 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800">
                            Open Browser
                        </button>
                    ) : null}
                    <button
                        onClick={handleFinish}
                        disabled={finishing || !authURL}
                        className="ml-3 inline-flex items-center text-white bg-slate-900 hover:bg-slate-800 focus:ring-4 focus:outline-none font-medium rounded-lg text-sm w-full sm:w-auto px-5 py-2.5 text-center dark:bg-green-600 dark:hover:bg-green-700 dark:focus:ring-green-800">
                        {finishing ? (
                            <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white"
                                 xmlns="http://www.w3.org/2000/svg"
                                 fill="none" viewBox="0 0 24 24">
                                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor"
                                        strokeWidth="4"/>
                                <path className="opacity-75" fill="currentColor"
                                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
                            </svg>
                        ) : null}
                        {finishing ? "Verifying..." : "I've Signed In"}
                    </button>
                </div>
            </div>
        </div>
    )
}

export default Auth;
