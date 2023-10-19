
# protocol
p=$1

go test -run="^BenchmarkAlea$" -v -bench=. -benchmem -benchtime 30s > ${p}.txt

cat ${p}.txt

mv profile.out ${p}.out

go tool pprof -pdf ${p}.out > ${p}.pdf
