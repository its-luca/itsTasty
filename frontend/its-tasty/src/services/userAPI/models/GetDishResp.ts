/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

/**
 * Detailed description of a dish
 */
export type GetDishResp = {
    /**
     * Name of the dish
     */
    name: string;
    /**
     * Location where this dish is served
     */
    servedAt: string;
    /**
     * Amount of times this dish occurred
     */
    occurrenceCount: number;
    /**
     * Most recent occurrences of the dish. Might not contain the whole history
     */
    recentOccurrences: Array<string>;
    /**
     * Average rating for this dish. Omitted if there are no votes yet
     */
    avgRating?: number;
    /**
     * Ratings for this dish. Keys mean rating, values mean ratings with that amount of stars. If more than zero votes are present avgRating field contains the average rating.
     */
    ratings: Record<string, number>;
    /**
     * Rating for this dish of the requesting user. Omitted if the user has not rated yet.
     */
    ratingOfUser?: GetDishResp.ratingOfUser;
};

export namespace GetDishResp {

    /**
     * Rating for this dish of the requesting user. Omitted if the user has not rated yet.
     */
    export enum ratingOfUser {
        '_1' = 1,
        '_2' = 2,
        '_3' = 3,
        '_4' = 4,
        '_5' = 5,
    }


}

