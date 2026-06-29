import { useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ingestDocument } from '../api/materias'
import './UploadPage.css'

type Status = 'idle' | 'uploading' | 'done' | 'error'

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
      const res = await ingestDocument(file, file.name, setProgress)
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
          disabled={status === 'uploading'}
        />
        {status === 'uploading' ? 'Enviando...' : 'Escolher arquivo PDF'}
      </label>

      {status === 'uploading' && (
        <div className="upload-page__progress" role="progressbar" aria-valuenow={progress}>
          <div className="upload-page__progress-bar" style={{ width: `${progress}%` }} />
          <span>{progress}%</span>
        </div>
      )}

      {status === 'error' && error && (
        <p className="upload-page__status upload-page__status--error">{error}</p>
      )}

      {status === 'done' && result && (
        <div className="upload-page__result">
          <p>
            {result.materiasCount} matéria(s) criada(s) a partir de {result.pages} página(s).
          </p>
          <button type="button" onClick={goToResults}>
            Ver matérias deste documento
          </button>
        </div>
      )}
    </div>
  )
}
