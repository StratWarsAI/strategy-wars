export * from './nav.type'
export * from './strategy.type'
export * from './simulation.type'
export * from './token.type'
export * from './websocket.type'
export * from './dashboard.type'

export interface SearchParams {
  [key: string]: string | string[] | undefined;
}