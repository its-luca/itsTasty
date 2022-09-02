/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

/**
 * Contains the dishID the requested dish
 */
export type SearchDishResp = {
    /**
     * True if the dish was found
     */
    foundDish: boolean;
    /**
     * ID of the searched dish if it was found. Omitted otherwise
     */
    dishID?: number;
    /**
     * Name of the searched ish
     */
    dishName: any;
};

