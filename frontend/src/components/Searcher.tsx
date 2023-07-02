import { h } from "dom-chef"
import { MetadataReponse, Language } from "../apitypes"


export default function Searcher(metadata: MetadataReponse, selectedLanguage: Language|null): JSX.Element {
    let languages = [
        {
            Name: "All languages",
            CountOfRepos: metadata.CountOfAllRepos,
            EscapedName: "",
        },
        ...metadata.Languages
    ]

    return (
        <>
            <div className="row m-3">
                <div className="col-md-3"></div>
                <div className="col-md-6">
                    <select id="searcher" className="selectpicker" data-live-search="true" data-width="100%">
                        {languages.map(language =>
                            <option
                                data-subtext={` ${language.CountOfRepos} ${language.CountOfRepos == 1 ? 'repo' : 'repos'}`}
                                selected={language == selectedLanguage}>

                                {language.Name ? language.Name : '(no language)'}
                            </option>
                        )}
                    </select>
                </div>
            </div>
        </>
    )
}