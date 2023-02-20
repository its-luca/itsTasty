/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

/**
 * Request to look up all dishes served on a date optionally filtered by a location
 */
export type SearchDishByDateReq = {
    /**
     * Date on which dishes must have been served. Format YYYY-MM-DD
     */
    date: string;
    /**
     * Location by which dishes must have been served
     */
    location?: string;
};

