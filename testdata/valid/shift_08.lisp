(defcolumns (X :i16))
;; intention is that shifts cancel.
(defconstraint c1 () (== 0 (- X (shift (shift X -1) 1 ))))
(defconstraint c2 () (== 0 (- (shift X 1) (shift (shift X -1) 2 ))))
