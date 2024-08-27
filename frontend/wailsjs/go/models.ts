export namespace main {
	
	export class ProjectData {
	    selected_directory: string;
	    file_list: {[key: string]: string[]};
	
	    static createFrom(source: any = {}) {
	        return new ProjectData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.selected_directory = source["selected_directory"];
	        this.file_list = source["file_list"];
	    }
	}

}

