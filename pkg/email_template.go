package pkg

import "fmt"

// PasswordResetTemplate generates the HTML body for a password reset email.
func PasswordResetTemplate(recipientName, resetURL string) string {
	return fmt.Sprintf(`
        <!DOCTYPE html>
        <html lang="en">
            <head>
              <meta charset="UTF-8" />
              <meta name="viewport" content="width=device-width, initial-scale=1.0" />
              <title>Password Reset Request</title>
              <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600&display=swap" rel="stylesheet">
            </head>
            <body style="margin:0; padding:0; background-color:#f4f6f8;">
              <table width="100%%" cellpadding="0" cellspacing="0" role="presentation" style="background-color:#f4f6f8;">
                <tr>
                  <td align="center" style="padding:40px 16px;">
                    <table width="100%%" cellpadding="0" cellspacing="0" role="presentation" style="max-width:600px; background-color:#ffffff; border-radius:8px; padding:32px; font-family:'Poppins', Arial, sans-serif; text-align:center; color:#333333;">

                      <tr>
                        <td style="font-size:22px; font-weight:600; padding-bottom:16px;">
                          Reset Password Anda
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:15px; padding-bottom:12px;">
                          Hai, <strong>%s</strong>
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:14px; line-height:1.6; padding-bottom:20px;">
                          Kami menerima permintaan untuk mereset password akun Anda di Yayasan Orang Tua Asuh.
                          <br /><br />
                          Untuk mereset password Anda, silakan klik tombol di bawah ini:
                        </td>
                      </tr>

                      <tr>
                        <td style="padding:20px 0;">
                          <a href="%s"
                             style="display:inline-block; background-color:#0E733B; color:#ffffff; text-decoration:none; font-size:14px; font-weight:500; padding:14px 96px; border-radius:6px;">
                            Reset Password
                          </a>
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:13px; line-height:1.6; padding-bottom:16px; color:#555555;">
                          Apabila tombol di atas tidak dapat diakses, Anda dapat mengklik atau menyalin dan menempelkan tautan di bawah ini di browser Anda:
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:13px; word-break:break-all; padding-bottom:20px;">
                          <a href="%s"
                             style="color:#2563eb; text-decoration:none;">
                            %s
                          </a>
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:13px; line-height:1.6; padding-bottom:24px; color:#555555;">
                          Link ini akan kedaluwarsa dalam 1 jam. Apabila Anda tidak merasa melakukan permintaan reset password, silakan abaikan email ini.
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:14px;">
                          Terima kasih,
                          <br />
                          <strong>Yayasan Orang Tua Asuh</strong>
                        </td>
                      </tr>

                    </table>
                  </td>
                </tr>
              </table>
            </body>
        </html>`, recipientName, resetURL, resetURL, resetURL)
}

// EmailVerificationTemplate generates the HTML body for email verification.
func EmailVerificationTemplate(recipientName, verificationURL string) string {
	return fmt.Sprintf(`
        <!DOCTYPE html>
        <html lang="id">
            <head>
              <meta charset="UTF-8" />
              <meta name="viewport" content="width=device-width, initial-scale=1.0" />
              <title>Verifikasi Email</title>
              <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600&display=swap" rel="stylesheet">
            </head>
            <body style="margin:0; padding:0; background-color:#f4f6f8;">
              <table width="100%%" cellpadding="0" cellspacing="0" role="presentation" style="background-color:#f4f6f8;">
                <tr>
                  <td align="center" style="padding:40px 16px;">
                    <table width="100%%" cellpadding="0" cellspacing="0" role="presentation" style="max-width:600px; background-color:#ffffff; border-radius:8px; padding:32px; font-family:'Poppins', Arial, sans-serif; text-align:center; color:#333333;">

                      <tr>
                        <td style="font-size:22px; font-weight:600; padding-bottom:16px;">
                          Verifikasi Alamat Email Anda
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:15px; padding-bottom:12px;">
                          Hai, <strong>%s</strong>
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:14px; line-height:1.6; padding-bottom:20px;">
                          Terima kasih telah melakukan pendaftaran akun pada Sistem Yayasan Orang Tua Asuh.
                          <br /><br />
                          Untuk mengaktifkan akun Anda, silakan melakukan verifikasi dengan menekan tombol di bawah ini:
                        </td>
                      </tr>

                      <tr>
                        <td style="padding:20px 0;">
                          <a href="%s"
                             style="display:inline-block; background-color:#0E733B; color:#ffffff; text-decoration:none; font-size:14px; font-weight:500; padding:14px 96px; border-radius:6px;">
                            Verifikasi Email
                          </a>
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:13px; line-height:1.6; padding-bottom:16px; color:#555555;">
                          Apabila tombol di atas tidak dapat diakses, Anda dapat mengklik atau menyalin dan menempelkan tautan di bawah ini di browser Anda:
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:13px; word-break:break-all; padding-bottom:20px;">
                          <a href="%s"
                             style="color:#2563eb; text-decoration:none;">
                            %s
                          </a>
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:13px; line-height:1.6; padding-bottom:24px; color:#555555;">
                          Apabila Anda tidak merasa melakukan pendaftaran akun, silakan abaikan email ini.
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:14px;">
                          Terima kasih,
                          <br />
                          <strong>Yayasan Orang Tua Asuh</strong>
                        </td>
                      </tr>

                    </table>
                  </td>
                </tr>
              </table>
            </body>
        </html>`, recipientName, verificationURL, verificationURL, verificationURL)
}

// FosterChildrenCandidateAcceptedTemplate generates the HTML body for a candidate acceptance email.
func FosterChildrenCandidateAcceptedTemplate(recipientName, candidateName string) string {
	return fmt.Sprintf(`
        <!DOCTYPE html>
        <html lang="id">
            <head>
              <meta charset="UTF-8" />
              <meta name="viewport" content="width=device-width, initial-scale=1.0" />
              <title>Calon Anak Asuh Disetujui</title>
              <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600&display=swap" rel="stylesheet">
            </head>
            <body style="margin:0; padding:0; background-color:#f4f6f8;">
              <table width="100%%" cellpadding="0" cellspacing="0" role="presentation" style="background-color:#f4f6f8;">
                <tr>
                  <td align="center" style="padding:40px 16px;">
                    <table width="100%%" cellpadding="0" cellspacing="0" role="presentation" style="max-width:600px; background-color:#ffffff; border-radius:8px; padding:32px; font-family:'Poppins', Arial, sans-serif; text-align:center; color:#333333;">

                      <tr>
                        <td style="font-size:22px; font-weight:600; padding-bottom:16px; color:#0E733B;">
                          Pendaftaran Disetujui
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:15px; padding-bottom:12px;">
                          Hai, <strong>%s</strong>
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:14px; line-height:1.6; padding-bottom:24px;">
                          Dengan senang hati kami mengumumkan bahwa pendaftaran calon anak asuh atas nama <strong>%s</strong> yang Anda ajukan telah <strong>disetujui</strong>.
                          <br /><br />
                          Calon anak asuh tersebut kini telah resmi terdaftar di Yayasan Orang Tua Asuh. Terima kasih atas partisipasi dan kepedulian Anda dalam membantu masa depan mereka.
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:14px; border-top:1px solid #eeeeee; padding-top:24px;">
                          Terima kasih,
                          <br />
                          <strong>Yayasan Orang Tua Asuh</strong>
                        </td>
                      </tr>

                    </table>
                  </td>
                </tr>
              </table>
            </body>
        </html>`, recipientName, candidateName)
}

// FosterChildrenCandidateRejectedTemplate generates the HTML body for a candidate rejection email.
func FosterChildrenCandidateRejectedTemplate(recipientName, candidateName, rejectionReason string) string {
	return fmt.Sprintf(`
        <!DOCTYPE html>
        <html lang="id">
            <head>
              <meta charset="UTF-8" />
              <meta name="viewport" content="width=device-width, initial-scale=1.0" />
              <title>Pendaftaran Calon Anak Asuh Ditolak</title>
              <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600&display=swap" rel="stylesheet">
            </head>
            <body style="margin:0; padding:0; background-color:#f4f6f8;">
              <table width="100%%" cellpadding="0" cellspacing="0" role="presentation" style="background-color:#f4f6f8;">
                <tr>
                  <td align="center" style="padding:40px 16px;">
                    <table width="100%%" cellpadding="0" cellspacing="0" role="presentation" style="max-width:600px; background-color:#ffffff; border-radius:8px; padding:32px; font-family:'Poppins', Arial, sans-serif; text-align:center; color:#333333;">

                      <tr>
                        <td style="font-size:22px; font-weight:600; padding-bottom:16px; color:#dc2626;">
                          Pendaftaran Belum Dapat Disetujui
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:15px; padding-bottom:12px;">
                          Hai, <strong>%s</strong>
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:14px; line-height:1.6; padding-bottom:20px;">
                          Terima kasih telah mengajukan calon anak asuh atas nama <strong>%s</strong> ke Yayasan Orang Tua Asuh.
                          <br /><br />
                          Setelah melalui proses evaluasi, dengan berat hati kami menginformasikan bahwa pendaftaran ini <strong>belum dapat disetujui</strong> karena alasan berikut:
                        </td>
                      </tr>

                      <tr>
                        <td align="left" style="background-color:#fef2f2; border-left:4px solid #dc2626; padding:16px; margin-bottom:24px; font-size:14px; border-radius:4px; line-height:1.5;">
                          <strong style="color:#dc2626;">Alasan Penolakan:</strong><br />
                          <span style="white-space: pre-wrap;">%s</span>
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:14px; padding-top:24px;">
                          Jika Anda memiliki pertanyaan lebih lanjut atau ingin mengajukan kembali dengan data yang disesuaikan, silakan hubungi kami.
                          <br /><br />
                          Terima kasih,
                          <br />
                          <strong>Yayasan Orang Tua Asuh</strong>
                        </td>
                  </td>
                </tr>
              </table>
            </body>
        </html>`, recipientName, candidateName, rejectionReason)
}

// AmbulanceServiceRequestAcceptedTemplate generates the HTML body for an ambulance request acceptance email.
func AmbulanceServiceRequestAcceptedTemplate(recipientName, submitterName string) string {
	return fmt.Sprintf(`
        <!DOCTYPE html>
        <html lang="id">
            <head>
              <meta charset="UTF-8" />
              <meta name="viewport" content="width=device-width, initial-scale=1.0" />
              <title>Permintaan Ambulans Disetujui</title>
              <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600&display=swap" rel="stylesheet">
            </head>
            <body style="margin:0; padding:0; background-color:#f4f6f8;">
              <table width="100%%" cellpadding="0" cellspacing="0" role="presentation" style="background-color:#f4f6f8;">
                <tr>
                  <td align="center" style="padding:40px 16px;">
                    <table width="100%%" cellpadding="0" cellspacing="0" role="presentation" style="max-width:600px; background-color:#ffffff; border-radius:8px; padding:32px; font-family:'Poppins', Arial, sans-serif; text-align:center; color:#333333;">

                      <tr>
                        <td style="font-size:22px; font-weight:600; padding-bottom:16px; color:#0E733B;">
                          Permintaan Ambulans Disetujui
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:15px; padding-bottom:12px;">
                          Hai, <strong>%s</strong>
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:14px; line-height:1.6; padding-bottom:24px;">
                          Permintaan layanan ambulans atas nama <strong>%s</strong> yang Anda ajukan telah <strong>disetujui</strong>.
                          <br /><br />
                          Tim kami akan segera mempersiapkan ambulans dan menghubungi Anda untuk koordinasi lebih lanjut. Harap pastikan nomor telepon Anda dapat dihubungi.
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:13px; line-height:1.6; padding-bottom:24px; color:#555555;">
                          Jika Anda memiliki pertanyaan, silakan hubungi kami segera.
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:14px; border-top:1px solid #eeeeee; padding-top:24px;">
                          Terima kasih,
                          <br />
                          <strong>Yayasan Orang Tua Asuh</strong>
                        </td>
                      </tr>

                    </table>
                  </td>
                </tr>
              </table>
            </body>
        </html>`, recipientName, submitterName)
}

// AmbulanceServiceRequestRejectedTemplate generates the HTML body for an ambulance request rejection email.
func AmbulanceServiceRequestRejectedTemplate(recipientName, submitterName, rejectionReason string) string {
	return fmt.Sprintf(`
        <!DOCTYPE html>
        <html lang="id">
            <head>
              <meta charset="UTF-8" />
              <meta name="viewport" content="width=device-width, initial-scale=1.0" />
              <title>Permintaan Ambulans Tidak Dapat Diproses</title>
              <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600&display=swap" rel="stylesheet">
            </head>
            <body style="margin:0; padding:0; background-color:#f4f6f8;">
              <table width="100%%" cellpadding="0" cellspacing="0" role="presentation" style="background-color:#f4f6f8;">
                <tr>
                  <td align="center" style="padding:40px 16px;">
                    <table width="100%%" cellpadding="0" cellspacing="0" role="presentation" style="max-width:600px; background-color:#ffffff; border-radius:8px; padding:32px; font-family:'Poppins', Arial, sans-serif; text-align:center; color:#333333;">

                      <tr>
                        <td style="font-size:22px; font-weight:600; padding-bottom:16px; color:#dc2626;">
                          Permintaan Ambulans Tidak Dapat Diproses
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:15px; padding-bottom:12px;">
                          Hai, <strong>%s</strong>
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:14px; line-height:1.6; padding-bottom:20px;">
                          Kami mohon maaf, permintaan layanan ambulans atas nama <strong>%s</strong> yang Anda ajukan <strong>tidak dapat diproses</strong> karena alasan berikut:
                        </td>
                      </tr>

                      <tr>
                        <td align="left" style="background-color:#fef2f2; border-left:4px solid #dc2626; padding:16px; margin-bottom:24px; font-size:14px; border-radius:4px; line-height:1.5;">
                          <strong style="color:#dc2626;">Alasan Penolakan:</strong><br />
                          <span style="white-space: pre-wrap;">%s</span>
                        </td>
                      </tr>

                      <tr>
                        <td style="font-size:14px; padding-top:24px;">
                          Jika Anda memiliki pertanyaan lebih lanjut atau ingin mengajukan permintaan kembali, silakan hubungi kami.
                          <br /><br />
                          Terima kasih,
                          <br />
                          <strong>Yayasan Orang Tua Asuh</strong>
                        </td>
                      </tr>

                    </table>
                  </td>
                </tr>
              </table>
            </body>
        </html>`, recipientName, submitterName, rejectionReason)
}
