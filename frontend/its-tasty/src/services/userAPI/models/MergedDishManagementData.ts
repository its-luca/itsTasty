/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

import type { ContainedDishEntry } from './ContainedDishEntry';

/**
 * Management Data for merged dish
 */
export type MergedDishManagementData = {
    /**
     * Location the merged dish is served at
     */
    servedAt: string;
    /**
     * Name of the merged dish
     */
    name: string;
    /**
     * Information about contained dishes
     */
    containedDishes: Array<ContainedDishEntry>;
};

