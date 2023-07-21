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
    FirstFetchedFromGithubAt: null | string
    FullName: string
    GithubLink: string
    Homepage: string
    Id: number
    Language: string
    LastFetchedFromGithubAt: string
    LicenseSpdxId: string
    Name: string
    OpenIssues: number
    OwnerAvatarUrl: string
    OwnerGravatarId: string
    OwnerLogin: string
    OwnerType: string
    RepoPushedAt: null | string
    RepoUpdatedAt: string
    Stargazers: number
    Topics: string
}
