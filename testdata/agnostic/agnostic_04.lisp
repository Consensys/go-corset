(defconst TWO_128 (^ 2 128))
;; construct u256 word from two u128 words
(defpurefun (to_u256 MSW LSW) (+ (* TWO_128 MSW) LSW))
;;
(module m1)
;;
(defcolumns (X_LO :i128) (X_HI :i128))
;;
(deflookup l1 (m2.Y) ((to_u256 X_HI X_LO)))

(module m2)
(defcolumns (Y :i256))
