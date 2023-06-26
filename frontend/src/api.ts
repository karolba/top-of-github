import { MetadataReponse, ToplistPageResponse } from './apitypes.js';

const BASE_API_URI = 'https://top-of-github-api.baraniecki.eu';

export async function getMetadata(): Promise<MetadataReponse> {
    let response = await fetch(`${BASE_API_URI}/metadata`);
    return await response.json();
}

export async function languageToplistPage(escapedLanguageName: string, pageNumber: number): Promise<ToplistPageResponse> {
    let response = await fetch(`${BASE_API_URI}/language/${escapedLanguageName}/${pageNumber}`);
    return await response.json();
}

export async function allLanguagesToplistPage(pageNumber: number): Promise<ToplistPageResponse> {
    let response = await fetch(`${BASE_API_URI}/all/${pageNumber}`);
    return await response.json();
}
