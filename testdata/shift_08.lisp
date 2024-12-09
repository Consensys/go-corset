(defcolumns X)
;; intention is that shifts cancel.
(defconstraint c1 () (- X (shift (shift X -1) 1 )))
(defconstraint c2 () (- (shift X 1) (shift (shift X -1) 2 )))
