#!/usr/bin/env bash 

function replace_llc_with_inc {
	sed -i 's/Gruntwork, LLC/Gruntwork, Inc/' LICENSE.txt
}

echo "Replacing Gruntwork, LLC with Gruntwork, Inc in LICENSE.txt file"

replace_llc_with_inc
