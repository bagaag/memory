#!/bin/bash
fn=`date '+%Y-%m-%d-%H-%M.zip'`
loc="$HOME/.memory/backups"
if [ ! -d "$loc" ] 
then
  mkdir $loc
fi
zip -r "$loc/$fn" ~/.memory/entries ~/.memory/files
