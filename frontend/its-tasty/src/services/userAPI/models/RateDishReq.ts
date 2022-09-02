/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

/**
 * Request to vote for a dish
 */
export type RateDishReq = {
    rating: RateDishReq.rating;
};

export namespace RateDishReq {

    export enum rating {
        '_1' = 1,
        '_2' = 2,
        '_3' = 3,
        '_4' = 4,
        '_5' = 5,
    }


}

