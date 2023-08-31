import { h } from "dom-chef"
import { Repository } from "../apitypes"
import Pagination from "./Pagination";
import { ScrollPositionSavingLink } from "../scrollPosition";

function addQueryParam(url: URL, param: string, value: string): URL {
    const params = new URLSearchParams(url.search);
  
    params.set(param, value);
  
    url.search = params.toString();
    return url;
}

function NotPushedToInAWhileBadge(lastPushedDaysAgo: number): JSX.Element {
    // TODO: replace this manual date math with something better
    let lastPushedYearsAgo = lastPushedDaysAgo / 365.25;
    
    if(lastPushedYearsAgo < 1)
        return <></>

    return <span className="badge bg-warning text-dark rounded-pill m-1">
        Last push {Math.round(lastPushedYearsAgo*10) / 10} years ago
    </span>
}

function CreatedAtBadge(createdAtDaysAgo: number): JSX.Element {
    // TODO: replace this manual date math with something better
    let createdAtYearsAgo = createdAtDaysAgo / 365.25;

    if(createdAtYearsAgo > 1) {
        return <span className="badge bg-secondary rounded-pill m-1">
            {Math.round(createdAtYearsAgo * 10) / 10} years old
        </span>
    } else if(createdAtDaysAgo > 31) {
        return <span className="badge bg-success rounded-pill m-1">
            {Math.round(createdAtYearsAgo * 12)} months old
        </span>
    } else {
        return <span className="badge bg-success rounded-pill m-1">
            {Math.round(createdAtDaysAgo)} days old
        </span>
    }

}

function githubAccountLink(accountName: string): string {
    // TODO: should probably fetch this from the API as well
    return `https://github.com/${encodeURIComponent(accountName)}`
}

function makeSureLinkIsntRelative(href: string): string {
    if(href.startsWith('https://')) {
        return href
    } else if(href.startsWith('http://')) {
        return href
    } else {
        return `http://${href}`
    }
}

function Repo(repo: Repository): JSX.Element {
    // TODO: replace this manual date math below with something better

    let lastPushedToBadge = <></>
    if(repo.RepoPushedAt) {
        let pushedAt = Date.parse(repo.RepoPushedAt)
        let pushedAtMillisecondsAgo = new Date().getTime() - pushedAt
        let pushedAtDaysAgo = pushedAtMillisecondsAgo / 1000 / 60 / 60 / 24
        lastPushedToBadge = NotPushedToInAWhileBadge(pushedAtDaysAgo)
    }

    let createdAtBadge = <></>
    if(repo.CreatedAt) {
        let createdAt = Date.parse(repo.CreatedAt)
        let createdAtMillisecondsAgo = new Date().getTime() - createdAt
        let createdAtDaysAgo = createdAtMillisecondsAgo / 1000 / 60 / 60 / 24
        createdAtBadge = CreatedAtBadge(createdAtDaysAgo)
    }

    const AVATAR_SIZE_PX = 50;

    // TODO: is there something better than LicenseSpdxId? A lot of repos show "NOASSERTION" despite
    // having a valid license
    return (
        <li className="list-group-item d-md-flex justify-content-between align-items-start">
            <div className="d-flex flex-column align-items-center">
                <b>{repo.Stargazers.toLocaleString('en')}</b>
                <small className="text-secondary">stargazers</small>
            </div>
            <div className="d-flex flex-column align-items-center ms-md-2">
                <img
                    width={AVATAR_SIZE_PX}
                    height={AVATAR_SIZE_PX}
                    src={addQueryParam(new URL(repo.OwnerAvatarUrl), 's', `${AVATAR_SIZE_PX * 2}`).toString()}
                    loading="lazy"
                ></img>
            </div>
            <div className="ms-md-2 w-100 text-break">
                <div className="d-md-flex">
                    <div className="me-auto" style={{overflowWrap: "anywhere"}}>
                        <ScrollPositionSavingLink href={githubAccountLink(repo.OwnerLogin)}>{repo.OwnerLogin}</ScrollPositionSavingLink>
                        <span className="user-select-none"> </span>
                        /
                        <span className="user-select-none"> </span>
                        <ScrollPositionSavingLink href={repo.GithubLink} className="fw-bold">{repo.Name}</ScrollPositionSavingLink>
                    </div>
                    <div className="text-md-end">
                        {repo.Language
                            ? <span className="badge bg-primary rounded-pill m-1">{repo.Language}</span>
                            : <></>
                        }
                        {repo.Archived
                            ? <span className="badge bg-warning text-dark rounded-pill m-1">Archived</span>
                            : <></>
                        }
                        {lastPushedToBadge}
                        {createdAtBadge}
                        {repo.LicenseSpdxId && repo.LicenseSpdxId != 'NOASSERTION'
                            ? <span className="badge bg-info text-dark rounded-pill m-1" title={repo.LicenseName}>{repo.LicenseSpdxId}</span>
                            : <></>
                        }
                    </div>
                </div>
                {repo.Description}
                {repo.Homepage
                    ? <> - <ScrollPositionSavingLink href={makeSureLinkIsntRelative(repo.Homepage)}>{repo.Homepage}</ScrollPositionSavingLink></>
                    : <></>
                }
            </div>
        </li>
    )
}

export default function Results(repositories: Repository[], page: number, pages: number, newPageHref: (page: number) => string): JSX.Element {
    return (
        <div>
            <ul className="list-group">
                {repositories.map(Repo)}
            </ul>
            {Pagination(page, pages, newPageHref)}
        </div>
    )
}