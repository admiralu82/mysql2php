<?php
header("Content-Type: text/json");

$DBHost = "localhost";
$DBLogin = "loginToDB";
$DBPassword = "passwordToDB";
$DBName = "nameOfDB"; 

$r = file_get_contents('php://input');
$out = json_decode($r,true);

$login = $out['User'];
$password = $out['Pass'];
$sql = $out['Sql'];

$link = mysqli_connect($DBHost,$DBLogin, $DBPassword, $DBName);
 
// Check connection
if($link === false){
    echo("ERROR: Could not connect. " . mysqli_connect_error());
    return;
}
 
if (($login!="MyLogin") && ($password!="MyPassword")) {
  echo("ERROR: Auth"); 
  return; 
}

if ($sql=='') {
  echo("ERROR: SQL"); 
  return ; 
}

///////////////////////////////////////////////
mysqli_query($link, "SET NAMES utf8");

$test_query = $sql;
//$test_query = mysqli_real_escape_string($link, $test_query);


$result = mysqli_query($link, $test_query);

$numRows = mysqli_affected_rows($link);

if($result===FALSE) {
  echo "ERROR SQL".$tablCnt." ".$test_query;
  return;
}

if($result===TRUE) {
  echo "OK ".$tablCnt;
  return;
}


echo "OK ".$numRows.PHP_EOL;

echo "[".PHP_EOL;

$tblCnt = 0;
$first = true;
while($tbl = mysqli_fetch_array($result)) {
  if ($first==false) {
    echo ",".PHP_EOL;
  }
  $first = false;
  $tblCnt++;
  //echo print_r($tbl,true);
  
  echo json_encode($tbl,JSON_UNESCAPED_UNICODE);  
}

echo PHP_EOL."]";

///////////////////////////////////////////////
// Print host information

 
// Close connection
mysqli_close($link);
?>