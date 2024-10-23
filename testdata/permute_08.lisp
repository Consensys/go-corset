(defcolumns
  (X :u16)
  (Y :u16))
(permute (A B) (+X -Y))
(defconstraint diag_ab () (* A (- (shift A 1) B)))
