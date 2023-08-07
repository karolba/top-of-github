export interface MetadataReponse {
    CountOfAllRepos: number
    CountOfAllStars: number
    AllReposPages: number
    Languages: Language[]
}

export interface Language {
    Name: string
    EscapedName: string
    CountOfRepos: number
    CountOfStars: number
    Pages: number
}

export type ToplistPageResponse = Repository[]

export interface Repository {
    Archived: number
    CreatedAt: string
    Description: string
    GithubLink: string
    Homepage: string
    Language: string
    LicenseSpdxId: string
    LicenseName: string
    Name: string
    OwnerAvatarUrl: string
    OwnerLogin: string
    RepoPushedAt: null | string
    RepoUpdatedAt: string
    Stargazers: number
}
