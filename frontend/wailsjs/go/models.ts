export namespace models {
	
	export class Detail {
	    id: number;
	    itemId: number;
	    text: string;
	    included: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Detail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.itemId = source["itemId"];
	        this.text = source["text"];
	        this.included = source["included"];
	    }
	}
	export class Item {
	    id: number;
	    sectionId: number;
	    data: Record<string, any>;
	    details: Detail[];
	    included: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Item(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sectionId = source["sectionId"];
	        this.data = source["data"];
	        this.details = this.convertValues(source["details"], Detail);
	        this.included = source["included"];
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
	export class PageLayout {
	    breaks: number[];
	    pageCount: number;
	
	    static createFrom(source: any = {}) {
	        return new PageLayout(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.breaks = source["breaks"];
	        this.pageCount = source["pageCount"];
	    }
	}
	export class Preset {
	    id: number;
	    name: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Preset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class Section {
	    id: number;
	    type: string;
	    title: string;
	    name: string;
	    headline: string;
	    titleColor: string;
	    lineColor: string;
	    visible: boolean;
	    items: Item[];
	
	    static createFrom(source: any = {}) {
	        return new Section(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.title = source["title"];
	        this.name = source["name"];
	        this.headline = source["headline"];
	        this.titleColor = source["titleColor"];
	        this.lineColor = source["lineColor"];
	        this.visible = source["visible"];
	        this.items = this.convertValues(source["items"], Item);
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
	export class TextStyle {
	    size: number;
	    bold: boolean;
	    italic: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TextStyle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.size = source["size"];
	        this.bold = source["bold"];
	        this.italic = source["italic"];
	    }
	}
	export class TextStyles {
	    sectionTitle: TextStyle;
	    entryTitle: TextStyle;
	    entryDates: TextStyle;
	    body: TextStyle;
	
	    static createFrom(source: any = {}) {
	        return new TextStyles(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sectionTitle = this.convertValues(source["sectionTitle"], TextStyle);
	        this.entryTitle = this.convertValues(source["entryTitle"], TextStyle);
	        this.entryDates = this.convertValues(source["entryDates"], TextStyle);
	        this.body = this.convertValues(source["body"], TextStyle);
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
	export class ResumeState {
	    styles: TextStyles;
	    presets: Preset[];
	    sections: Section[];
	
	    static createFrom(source: any = {}) {
	        return new ResumeState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.styles = this.convertValues(source["styles"], TextStyles);
	        this.presets = this.convertValues(source["presets"], Preset);
	        this.sections = this.convertValues(source["sections"], Section);
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

