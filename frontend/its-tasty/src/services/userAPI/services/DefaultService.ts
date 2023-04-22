/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { CreateMergedDishReq } from '../models/CreateMergedDishReq';
import type { CreateMergedDishResp } from '../models/CreateMergedDishResp';
import type { GetAllDishesResponse } from '../models/GetAllDishesResponse';
import type { GetDishResp } from '../models/GetDishResp';
import type { GetMergeCandidatesResp } from '../models/GetMergeCandidatesResp';
import type { GetUsersMeResp } from '../models/GetUsersMeResp';
import type { MergedDishManagementData } from '../models/MergedDishManagementData';
import type { MergedDishUpdateReq } from '../models/MergedDishUpdateReq';
import type { RateDishReq } from '../models/RateDishReq';
import type { SearchDishByDateReq } from '../models/SearchDishByDateReq';
import type { SearchDishReq } from '../models/SearchDishReq';
import type { SearchDishResp } from '../models/SearchDishResp';

import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';

export class DefaultService {

    /**
     * Get information about the user doing this requests
     * @returns GetUsersMeResp Valid session. Return info about user
     * @throws ApiError
     */
    public static getUsersMe(): CancelablePromise<GetUsersMeResp> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/users/me',
            errors: {
                401: `User needs to login`,
                500: `Internal error but input was fine`,
            },
        });
    }

    /**
     * Returns the IDs of all known dishes
     * @returns GetAllDishesResponse Return all known dish IDs
     * @throws ApiError
     */
    public static getGetAllDishes(): CancelablePromise<GetAllDishesResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/getAllDishes',
            errors: {
                401: `User needs to login`,
                500: `Internal error but input was fine`,
            },
        });
    }

    /**
     * Search for a dish by name
     * @param requestBody
     * @returns SearchDishResp Success
     * @throws ApiError
     */
    public static postSearchDish(
        requestBody: SearchDishReq,
    ): CancelablePromise<SearchDishResp> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/searchDish',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                401: `User needs to login`,
                500: `Internal error but input was fine`,
            },
        });
    }

    /**
     * Search for a dish by Date and optional by location
     * @param requestBody
     * @returns number Success. Array with matching dish ids (may be empty)
     * @throws ApiError
     */
    public static postSearchDishByDate(
        requestBody: SearchDishByDateReq,
    ): CancelablePromise<Array<number>> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/searchDish/byDate',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                401: `User needs to login`,
                500: `Internal error but input was fine`,
            },
        });
    }

    /**
     * Returns dishes that have a similar name and should probably be merged with this dish
     * @param dishId
     * @returns GetMergeCandidatesResp Success
     * @throws ApiError
     */
    public static getDishesMergeCandidates(
        dishId: number,
    ): CancelablePromise<GetMergeCandidatesResp> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/dishes/mergeCandidates/{dishID}',
            path: {
                'dishID': dishId,
            },
            errors: {
                401: `User needs to login`,
                404: `dishID not found`,
                500: `Internal error but input was fine`,
            },
        });
    }

    /**
     * Rate the dish.
     * @param dishId
     * @param requestBody
     * @returns any Success
     * @throws ApiError
     */
    public static postDishes(
        dishId: number,
        requestBody: RateDishReq,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/dishes/{dishID}',
            path: {
                'dishID': dishId,
            },
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Bad Input Data`,
                401: `User needs to login`,
                404: `dishID not found`,
                500: `Internal error but input was fine`,
            },
        });
    }

    /**
     * Get details like ratings and occurrences for this dish including the users own rating. If this dish \ is part of a merged dish, we return the data for the merged dish instead of the individual dish
     * @param dishId
     * @returns GetDishResp Detailed information about the dish
     * @throws ApiError
     */
    public static getDishes(
        dishId: number,
    ): CancelablePromise<GetDishResp> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/dishes/{dishID}',
            path: {
                'dishID': dishId,
            },
            errors: {
                400: `Bad Input data`,
                401: `User needs to login`,
                404: `dishID not found`,
                500: `Internal error but input was fine`,
            },
        });
    }

    /**
     * Create a new merged dish
     * @param requestBody
     * @returns CreateMergedDishResp Success. Merged dish was created
     * @throws ApiError
     */
    public static postMergedDishes(
        requestBody: CreateMergedDishReq,
    ): CancelablePromise<CreateMergedDishResp> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/mergedDishes/',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Bad Input Data. See error message`,
                401: `User needs to login`,
                500: `Internal server error but input was fine`,
            },
        });
    }

    /**
     * Get metadata for merged dish. Just for managing the merged dish object. Use /dishes/ endoints \ to get ratings etc.
     * @param mergedDishId
     * @returns MergedDishManagementData Success
     * @throws ApiError
     */
    public static getMergedDishes(
        mergedDishId: number,
    ): CancelablePromise<MergedDishManagementData> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/mergedDishes/{mergedDishID}',
            path: {
                'mergedDishID': mergedDishId,
            },
            errors: {
                401: `User needs to login`,
                404: `Merged dish not found`,
                500: `Internal server error but input was fine`,
            },
        });
    }

    /**
     * Update the values of the merged dish
     * @param mergedDishId
     * @param requestBody
     * @returns any Success. Merged dish was updated
     * @throws ApiError
     */
    public static patchMergedDishes(
        mergedDishId: number,
        requestBody: MergedDishUpdateReq,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'PATCH',
            url: '/mergedDishes/{mergedDishID}',
            path: {
                'mergedDishID': mergedDishId,
            },
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Bad Input Data. See error message`,
                401: `User needs to login`,
                404: `Merged dish not found`,
                500: `Internal server error but input was fine`,
            },
        });
    }

    /**
     * Delete the merged dish
     * @param mergedDishId
     * @returns any Success. Merged dish was deleted
     * @throws ApiError
     */
    public static deleteMergedDishes(
        mergedDishId: number,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/mergedDishes/{mergedDishID}',
            path: {
                'mergedDishID': mergedDishId,
            },
            errors: {
                400: `Bad Input Data. See error message`,
                401: `User needs to login`,
                404: `Merged dish not found`,
                500: `Internal server error but input was fine`,
            },
        });
    }

}
