(module sq)
(defcolumns (X_lo :i128) (X_hi :i128) (Y_lo :i128) (Y_hi :i128))
;; Y = (X * X) % 2^256
(deflookup l1 (add.arg1 add.arg2 add.res) (X_hi::X_lo X_hi::X_lo Y_hi::Y_lo))
