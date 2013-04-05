// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Copyright 2011 The Snappy-Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the SNAPPY-GO-LICENSE file.

/*
Package zappy implements the zappy block-based compression format.  It aims for
a combination of good speed and reasonable compression.

Zappy is a format incompatible, API compatible fork of snappy-go[1]. The C++
snappy implementation is at [2].

Reasons for the fork

The snappy compression is pretty good. Yet it has one problem built into its
format definition[3] - the maximum length of a copy "instruction" is 64 bytes.
For specific usage pattern with long runs of repeated data, it turns out the
compression is suboptimal. For example a 1:1000 "sparseness" 64kB bit index
with only few set bits is compressed to about 3kB (about 1000 of 64B copy, 3
byte "instructions").

Format description

Zappy uses much less complicated format than snappy. Each encoded block begins
with the uvarint-encoded[4] length of the decoded data, followed by a sequence
of chunks. Chunks begin and end on byte boundaries. The chunk starts with a
varint encoded number N:

	N >= 0: N+1 literal bytes follow.

	N < 0: copy -N bytes, starting at offset M (in the following uvarint).

Performance issues

Compression rate is roughly the same as of snappy for the reference data set:

                  testdata/html: snappy   23320, zappy   22943, 0.984, orig  102400
              testdata/urls.10K: snappy  334437, zappy  355163, 1.062, orig  702087
             testdata/house.jpg: snappy  126711, zappy  126694, 1.000, orig  126958
  testdata/mapreduce-osdi-1.pdf: snappy   77227, zappy   77646, 1.005, orig   94330
              testdata/html_x_4: snappy   92350, zappy   22956, 0.249, orig  409600
               testdata/cp.html: snappy   11938, zappy   12961, 1.086, orig   24603
              testdata/fields.c: snappy    4825, zappy    5395, 1.118, orig   11150
           testdata/grammar.lsp: snappy    1814, zappy    1933, 1.066, orig    3721
           testdata/kennedy.xls: snappy  423518, zappy  440597, 1.040, orig 1029744
           testdata/alice29.txt: snappy   89550, zappy  104016, 1.162, orig  152089
          testdata/asyoulik.txt: snappy   79583, zappy   91345, 1.148, orig  125179
            testdata/lcet10.txt: snappy  238761, zappy  275488, 1.154, orig  426754
          testdata/plrabn12.txt: snappy  324567, zappy  376885, 1.161, orig  481861
                  testdata/ptt5: snappy   96350, zappy   91465, 0.949, orig  513216
                   testdata/sum: snappy   18927, zappy   20015, 1.057, orig   38240
               testdata/xargs.1: snappy    2532, zappy    2793, 1.103, orig    4227
         testdata/geo.protodata: snappy   23362, zappy   20759, 0.889, orig  118588
             testdata/kppkn.gtb: snappy   73962, zappy   87200, 1.179, orig  184320
                          TOTAL: snappy 2043734, zappy 2136254, 1.045, orig 4549067

Zappy has better RLE handling (1/1000+1 non zero bytes in each index):

     Sparse bit index      16 B: snappy       9, zappy       9, 1.000
     Sparse bit index      32 B: snappy      10, zappy      10, 1.000
     Sparse bit index      64 B: snappy      11, zappy      10, 0.909
     Sparse bit index     128 B: snappy      16, zappy      14, 0.875
     Sparse bit index     256 B: snappy      22, zappy      14, 0.636
     Sparse bit index     512 B: snappy      36, zappy      16, 0.444
     Sparse bit index    1024 B: snappy      57, zappy      18, 0.316
     Sparse bit index    2048 B: snappy     111, zappy      32, 0.288
     Sparse bit index    4096 B: snappy     210, zappy      31, 0.148
     Sparse bit index    8192 B: snappy     419, zappy      75, 0.179
     Sparse bit index   16384 B: snappy     821, zappy     138, 0.168
     Sparse bit index   32768 B: snappy    1627, zappy     232, 0.143
     Sparse bit index   65536 B: snappy    3243, zappy     451, 0.139

Zappy _is_ slower than snappy. No profiling was attempted yet. Old=snappy,
new=zappy:

  benchmark                   old MB/s     new MB/s  speedup
  BenchmarkWordsDecode1e3       147.58       104.54    0.71x
  BenchmarkWordsDecode1e4       148.82       101.84    0.68x
  BenchmarkWordsDecode1e5       144.15        96.65    0.67x
  BenchmarkWordsDecode1e6       165.54       105.35    0.64x
  BenchmarkWordsEncode1e3        44.87        58.01    1.29x
  BenchmarkWordsEncode1e4        76.43        68.93    0.90x
  BenchmarkWordsEncode1e5        74.36        66.07    0.89x
  BenchmarkWordsEncode1e6        92.48        77.75    0.84x
  Benchmark_UFlat0              315.39       235.37    0.75x
  Benchmark_UFlat1              231.21       174.24    0.75x
  Benchmark_UFlat2             3648.39      3545.97    0.97x
  Benchmark_UFlat3              888.99       653.00    0.73x
  Benchmark_UFlat4              314.35       388.09    1.23x
  Benchmark_UFlat5              210.82       155.55    0.74x
  Benchmark_UFlat6              210.01       143.97    0.69x
  Benchmark_UFlat7              209.67       147.15    0.70x
  Benchmark_UFlat8              246.16       154.36    0.63x
  Benchmark_UFlat9              162.17       107.38    0.66x
  Benchmark_UFlat10             154.35       102.46    0.66x
  Benchmark_UFlat11             168.85       112.04    0.66x
  Benchmark_UFlat12             146.92        95.42    0.65x
  Benchmark_UFlat13             358.05       274.91    0.77x
  Benchmark_UFlat14             196.14       143.42    0.73x
  Benchmark_UFlat15             185.66       132.46    0.71x
  Benchmark_UFlat16             361.95       283.05    0.78x
  Benchmark_UFlat17             221.22       137.61    0.62x
  Benchmark_ZFlat0              187.55       171.16    0.91x
  Benchmark_ZFlat1               93.09        99.12    1.06x
  Benchmark_ZFlat2               67.69        98.88    1.46x
  Benchmark_ZFlat3               78.10       105.44    1.35x
  Benchmark_ZFlat4              190.93       336.84    1.76x
  Benchmark_ZFlat5               97.11        88.57    0.91x
  Benchmark_ZFlat6              105.53        96.97    0.92x
  Benchmark_ZFlat7               84.50        89.28    1.06x
  Benchmark_ZFlat8              151.11       123.57    0.82x
  Benchmark_ZFlat9               83.27        72.59    0.87x
  Benchmark_ZFlat10              77.84        67.99    0.87x
  Benchmark_ZFlat11              87.23        76.18    0.87x
  Benchmark_ZFlat12              75.02        64.33    0.86x
  Benchmark_ZFlat13             216.06       202.36    0.94x
  Benchmark_ZFlat14              93.19        86.37    0.93x
  Benchmark_ZFlat15              74.34        78.06    1.05x
  Benchmark_ZFlat16             212.14       203.21    0.96x
  Benchmark_ZFlat17             134.72       112.39    0.83x

Information sources

... referenced from the above documentation.

 [1]: http://code.google.com/p/snappy-go/
 [2]: http://code.google.com/p/snappy/
 [3]: http://code.google.com/p/snappy/source/browse/trunk/format_description.txt
 [4]: http://golang.org/pkg/encoding/binary/
*/
package zappy
