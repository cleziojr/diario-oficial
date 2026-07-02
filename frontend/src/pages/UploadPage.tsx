import { useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ingestDocument } from '../api/materias'
import './UploadPage.css'

type Status = 'idle' | 'uploading' | 'processing' | 'done' | 'error'

const busy = (s: Status) => s === 'uploading' || s === 'processing'

export function UploadPage() {
  const [status, setStatus] = useState<Status>('idle')
  const [progress, setProgress] = useState(0)
  const [error, setError] = useState<string | null>(null)
  const [result, setResult] = useState<{ materiasCount: number; pages: number } | null>(null)
  const [documentId, setDocumentId] = useState<string | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)
  const navigate = useNavigate()

  async function handleFile(file: File) {
    setStatus('uploading')
    setProgress(0)
    setError(null)
    setResult(null)
    try {
      const res = await ingestDocument(file, file.name, setProgress, () => setStatus('processing'))
      setResult({ materiasCount: res.materias_count, pages: res.pages })
      setDocumentId(res.document_id)
      setStatus('done')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao enviar o arquivo')
      setStatus('error')
    }
  }

  function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (file) void handleFile(file)
    // permite reenviar o mesmo arquivo depois
    e.target.value = ''
  }

  function goToResults() {
    if (!documentId) return
    navigate(`/?document_id=${encodeURIComponent(documentId)}`)
  }

  return (
    <div className="upload-page">
      <h1>Enviar Diário Oficial (PDF)</h1>
      <p>O documento será processado e suas matérias indexadas automaticamente.</p>

      <label className="upload-page__drop">
        <input
          ref={inputRef}
          type="file"
          accept="application/pdf"
          onChange={handleChange}
          disabled={busy(status)}
        />
        {status === 'uploading'
          ? 'Enviando...'
          : status === 'processing'
            ? 'Processando...'
            : 'Escolher arquivo PDF'}
      </label>

      {status === 'uploading' && (
        <div className="upload-page__progress" role="progressbar" aria-valuenow={progress}>
          <div className="upload-page__progress-bar" style={{ width: `${progress}%` }} />
          <span>Enviando arquivo… {progress}%</span>
        </div>
      )}

      {status === 'processing' && (
        <div className="upload-page__processing" role="status" aria-live="polite">
          <span className="upload-page__spinner" aria-hidden="true" />
          <span>
            Analisando o diário com IA e categorizando as matérias…
            <br />
            <small>Isso pode levar até 1 minuto. Não feche esta página.</small>
          </span>
        </div>
      )}

      {status === 'error' && error && (
        <div className="upload-page__status upload-page__status--error" role="alert">
          <p>{error}</p>
          <small>
            Se foi um erro de limite da IA (429), aguarde alguns segundos e tente novamente.
          </small>
        </div>
      )}

      {status === 'done' && result && (
        <div className="upload-page__result" role="status">
          <p>
            ✅ {result.materiasCount} matéria(s) criada(s) a partir de {result.pages} página(s).
          </p>
          <button type="button" onClick={goToResults}>
            Ver matérias deste documento
          </button>
        </div>
      )}
    </div>
  )
}
