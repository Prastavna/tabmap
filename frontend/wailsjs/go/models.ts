export namespace main {
	
	export class Button {
	    number: number;
	    currentAction: string;
	    isConfigurable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Button(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.number = source["number"];
	        this.currentAction = source["currentAction"];
	        this.isConfigurable = source["isConfigurable"];
	    }
	}
	export class Device {
	    name: string;
	    id: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new Device(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.id = source["id"];
	        this.type = source["type"];
	    }
	}
	export class DeviceMappings {
	    buttons: Record<number, string>;
	    wheels: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new DeviceMappings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.buttons = source["buttons"];
	        this.wheels = source["wheels"];
	    }
	}
	export class WheelAction {
	    property: string;
	    label: string;
	    currentAction: string;
	
	    static createFrom(source: any = {}) {
	        return new WheelAction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.property = source["property"];
	        this.label = source["label"];
	        this.currentAction = source["currentAction"];
	    }
	}

}

