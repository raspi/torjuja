/* Do not change, this code is generated from Golang structs */


export class AllowDTO {
    fqdn: string;

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.fqdn = source["fqdn"];
    }
}