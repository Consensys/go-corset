(defconst TWO_128 (^ 2 128))
;; construct u256 word from two u128 words
(defpurefun (to_u256 MSW LSW) (+ (* TWO_128 MSW) LSW))
;;
(defcolumns (X_LO :i128) (X_HI :i128) (Y :i256))
;;
(defconstraint c1 () (== Y (to_u256 X_HI X_LO)))
