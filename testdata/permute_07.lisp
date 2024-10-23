(defcolumns
  (ST :u16)
  (X :u16)
  (Y :u16))
(permute (ST' A B) (+ST -X +Y))
(defconstraint diag_ab () (* ST' (- (shift A 1) B)))
