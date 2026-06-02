package com.sshcustom.app.service

import android.content.Context
import java.io.File
import java.io.FileInputStream
import java.io.FileOutputStream

class SSHControlService(private val context: Context) {

    private val workDir = "/data/adb/sshcustom"
    private val sockPath = "$workDir/run/sshcustomd.sock"
    private val serviceScript = "$workDir/scripts/ssh.service"
    private val settingsPath = "$workDir/settings.ini"

    fun startTunnel() {
        exec("sh $serviceScript start")
    }

    fun stopTunnel() {
        exec("sh $serviceScript stop")
    }

    fun restartTunnel() {
        exec("sh $serviceScript restart")
    }

    fun isRunning(): Boolean {
        return exec("sh $serviceScript status").contains("running")
    }

    fun writeSetting(key: String, value: String) {
        exec("sed -i 's|^$key=.*|$key=\"$value\"|' $settingsPath")
    }

    private fun exec(cmd: String): String {
        return try {
            val process = Runtime.getRuntime().exec(arrayOf("su", "-c", cmd))
            process.inputStream.bufferedReader().readText()
        } catch (e: Exception) {
            ""
        }
    }
}