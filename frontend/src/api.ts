import { MetadataReponse, ToplistPageResponse } from './apitypes.js';

const BASE_API_URI = 'https://top-of-github-api.baraniecki.eu';

export async function getMetadata(): Promise<MetadataReponse> {
    let response = await fetch(`${BASE_API_URI}/metadata.json.gz`);
    return await response.json();
}

export async function languageToplistPage(escapedLanguageName: string, pageNumber: number): Promise<ToplistPageResponse> {
    let response = await fetch(`${BASE_API_URI}/language/${escapedLanguageName}/${pageNumber}.json.gz`);
    return await response.json();
}

export async function allLanguagesToplistPage(pageNumber: number): Promise<ToplistPageResponse> {
    let response = await fetch(`${BASE_API_URI}/all/${pageNumber}.json.gz`);
    return await response.json();
}
