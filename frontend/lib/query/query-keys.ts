export class QueryKeys<T extends string> {
    constructor(private readonly entity: T) {}
  
    all = () => [this.entity] as const;
    detail = (id: string) => [...this.all(), id] as const;
}