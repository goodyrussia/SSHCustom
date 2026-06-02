package com.sshcustom.app.service

import android.app.Service
import android.content.Intent
import android.os.Binder
import android.os.IBinder

class SSHControlService : Service() {

    inner class LocalBinder : Binder() {
        fun getService(): SSHControlService = this@SSHControlService
        fun startTunnel() = this@SSHControlService.startTunnel()
        fun stopTunnel() = this@SSHControlService.stopTunnel()
        fun restartTunnel() = this@SSHControlService.restartTunnel()
        fun isRunning() = this@SSHControlService.isRunning()
        fun readLogFile(path: String, lines: Int) = this@SSHControlService.readLogFile(path, lines)
        fun clearLogFile(path: String) = this@SSHControlService.clearLogFile(path)
        fun writeSettings(pairs: Map<String, String>) = this@SSHControlService.writeSettings(pairs)
        fun writeSetting(key: String, value: String) = this@SSHControlService.writeSetting(key, value)
    }

    private val binder = LocalBinder()
    private val workDir = "/data/adb/sshcustom"
    private val serviceScript = "$workDir/scripts/ssh.service"
    private val settingsPath = "$workDir/settings.ini"

    override fun onBind(intent: Intent?): IBinder = binder

    fun startTunnel() { rootExec("sh $serviceScript start") }
    fun stopTunnel() { rootExec("sh $serviceScript stop") }
    fun restartTunnel() { rootExec("sh $serviceScript restart") }

    fun isRunning(): Boolean {
        return rootExec("sh $serviceScript status").contains("running")
    }

    fun readLogFile(path: String, lines: Int): String {
        return rootExec("tail -n $lines '$path' 2>/dev/null || echo '(empty)'")
    }

    fun clearLogFile(path: String) { rootExec(": > '$path'") }

    fun writeSettings(pairs: Map<String, String>) {
        pairs.forEach { (key, value) ->
            rootExec("sed -i 's|^$key=.*|$key=\"$value\"|' $settingsPath")
        }
    }

    fun writeSetting(key: String, value: String) {
        rootExec("sed -i 's|^$key=.*|$key=\"$value\"|' $settingsPath")
    }

    private fun rootExec(cmd: String): String {
        return try {
            val process = Runtime.getRuntime().exec(arrayOf("su", "-c", cmd))
            process.inputStream.bufferedReader().readText().trim()
        } catch (e: Exception) { "" }
    }
}