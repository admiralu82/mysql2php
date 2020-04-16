<?php
$remote = $_SERVER['REMOTE_ADDR'];
$port = 4000;


header('Content-Type: text/html; charset=utf-8');
echo 'Привет, ' . $_GET["ID"] . " FROM " . $_SERVER['REMOTE_ADDR'] . "<br/>" . PHP_EOL;

$sockSql = socket_create(AF_INET, SOCK_STREAM, SOL_TCP);
$sockMaster = socket_create(AF_INET, SOCK_STREAM, SOL_TCP);

socket_set_option($sockSql, SOL_SOCKET, SO_REUSEADDR, 1);
socket_set_option($sockMaster, SOL_SOCKET, SO_REUSEADDR, 1);

if (socket_connect($sockSql, "127.0.0.1", 3306) == false) {
    $errorcode = socket_last_error();
    $errormsg = socket_strerror($errorcode);

    echo ("Couldn't create SQL socket: [$errorcode] $errormsg" . "<br/>");
    return;
}
;


if (socket_connect($sockMaster, $remote, $port) == false) {
    $errorcode = socket_last_error();
    $errormsg = socket_strerror($errorcode);

    echo ("Couldn't create master socket: [$errorcode] $errormsg" . "<br/>" . PHP_EOL);
    return;
}
;

// запишем ID
socket_write($sockMaster, chr((int) $_GET["ID"]));
echo ("Sent ID<br/>" . PHP_EOL);
$toSqlBytes = $toMasterBytes = 0;
while (true) {

    $readArr = array($sockSql, $sockMaster);
    $writeArr = array();
    $excpArr = array($sockSql, $sockMaster);

    

    if (socket_select($readArr, $writeArr, $excpArr, null) < 1) {
        continue;
    };

    if (count($excpArr)>0) {
        echo "Exception. Exit.<br/>" . PHP_EOL;
        break;
    }
    
    if (in_array($sockSql, $readArr) == true) {
        $fromSql = socket_read($sockSql, 10000);
        if ($fromSql === false) {
            break;
        }
        $l = socket_write($sockMaster, $fromSql);
        if ($l === false) {
            break;
        }
        $toMasterBytes += $l;
        //echo "ToMaster " . $l . " bytes<br/>" . PHP_EOL;
    }

    if (in_array($sockMaster, $readArr) == true) {
        $fromMain = socket_read($sockMaster, 10000);
        if ($fromMain === false) {
            break;
        }
        $l = socket_write($sockSql, $fromMain);
        if ($l === false) {
            break;
        }
        $toSqlBytes += $l;
        //echo "To   SQL " . $l . " bytes<br/>" . PHP_EOL;
    }
}

echo "Exit. ToSql writen ".$toSqlBytes. " bytes. ToMaster writen ".  $toMasterBytes. " bytes." . "<br/>" . PHP_EOL;
