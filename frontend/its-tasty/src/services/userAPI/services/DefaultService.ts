/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { GetAllDishesResponse } from '../models/GetAllDishesResponse';
import type { GetDishResp } from '../models/GetDishResp';
import type { GetUsersMeResp } from '../models/GetUsersMeResp';
import type { RateDishReq } from '../models/RateDishReq';
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
     * Rate the dish
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
     * Get details like ratings and occurrences for this dish including the users own rating
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

}
