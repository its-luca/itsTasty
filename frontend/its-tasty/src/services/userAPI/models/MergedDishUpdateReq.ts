/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

/**
 * Representation of a merged dish
 */
export type MergedDishUpdateReq = {
    /**
     * If present, these IDs are added to the merged dish.
     */
    addDishIDs?: Array<number>;
    /**
     * If present, these IDs are removed from the merged dish. At least two dish must remain. \ To delete a merge dish, use DELETE instead of PATCH
     */
    removeDishIDs?: Array<number>;
    /**
     * If present, the merged dish will be renamed to this
     */
    name?: string;
};

