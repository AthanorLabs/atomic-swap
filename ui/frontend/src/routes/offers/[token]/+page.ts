import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import { offers, selectedOffer, refreshOffers } from '../../../stores/offerStore'
	
export const load = (async ({ params }) => {
    return {
        token: params.token.toUpperCase()
    }
}) satisfies PageLoad;