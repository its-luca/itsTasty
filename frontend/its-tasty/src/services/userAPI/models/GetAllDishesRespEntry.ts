/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

/**
 * Entry in the result array returned by GetAllDishesResponse
 */
export type GetAllDishesRespEntry = {
    /**
     * dishID
     */
    id: number;
    /**
     * Optional field, if this dish is part of a merged dish
     */
    mergedDishID?: number;
    /**
     * Name of this dish
     */
    name: string;
    /**
     * Location where this dish is served at
     */
    servedAt: string;
};

