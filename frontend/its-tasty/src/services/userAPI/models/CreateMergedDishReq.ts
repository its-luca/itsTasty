/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

/**
 * Request to create a new MergedDish
 */
export type CreateMergedDishReq = {
    /**
     * Name of the merged dish. May be equal to existing dishes but for a given \ location there may not be another merged dish with the same name.
     */
    name: string;
    /**
     * Array of dish ids that should be merged. All dishes must be served at the same location \ and cannot be part of any other merged dishes.  At least two dishes must be provided
     */
    mergedDishes: Array<number>;
};

