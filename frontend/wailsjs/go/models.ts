export namespace models {
	
	export class RetryLevel {
	    query_suffix: string;
	    tolerance: number;
	
	    static createFrom(source: any = {}) {
	        return new RetryLevel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.query_suffix = source["query_suffix"];
	        this.tolerance = source["tolerance"];
	    }
	}
	export class Config {
	    download_path: string;
	    workers: number;
	    retries: RetryLevel[];
	    debug_mode: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.download_path = source["download_path"];
	        this.workers = source["workers"];
	        this.retries = this.convertValues(source["retries"], RetryLevel);
	        this.debug_mode = source["debug_mode"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

