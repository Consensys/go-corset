(defcolumns
  (X :u8)
  (Y :u8))
(permute (A B) (+X +Y))
(vanish diag_ab (- (shift A 1) B))
