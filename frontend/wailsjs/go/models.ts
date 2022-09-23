export namespace main {
	
	export class ProxyInfo {
	    remote_address: string;
	    local_address: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.remote_address = source["remote_address"];
	        this.local_address = source["local_address"];
	    }
	}

}

