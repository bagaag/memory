#!/bin/bash
MEMHOME="$HOME/.memory"
FN=`date '+%Y-%m-%d-%H-%M.zip'`
BU="$MEMHOME/backups"
if [ ! -d "$BU" ] 
then
  mkdir $BU
fi
cd $MEMHOME
zip -r "$BU/$FN" entries files
