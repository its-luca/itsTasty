/**
 * Contains defintions for local storage keys used throughout the app
 */

//we need this prefix to enable hosting multiple instances on subfolders on the same domain without a collision
const keyPrefix = (new URL(process.env.REACT_APP_PUBLIC_URL!).pathname).replace("/","_")

export const lsKeyIsAuthenticated = keyPrefix+"_isAuthenticated"