(module sq)
(defcolumns (X'0 :i128) (X'1 :i128) (Y'0 :i128) (Y'1 :i128))
;; Y = (X * X) % 2^256
(deflookup l1 (add.arg1'0 add.arg1'1 add.arg2'0 add.arg2'1 add.res'0 add.res'1) (X'0 X'1 X'0 X'1 Y'0 Y'1))
