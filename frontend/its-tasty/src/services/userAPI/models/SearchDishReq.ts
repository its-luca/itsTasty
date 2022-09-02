/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

/**
 * Request to lookup a dishID by the dish name
 */
export type SearchDishReq = {
    /**
     * Dish to search for
     */
    dishName: string;
    /**
     * Name of the location where this dish is served
     */
    servedAt: string;
};

