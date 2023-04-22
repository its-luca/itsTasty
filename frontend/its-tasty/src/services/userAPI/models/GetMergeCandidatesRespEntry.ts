/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

export type GetMergeCandidatesRespEntry = {
    /**
     * dish ID
     */
    dishID: number;
    /**
     * dish name
     */
    dishName: string;
    /**
     * If set, this dish is already part of a merged dish
     */
    mergedDishID?: number;
};

