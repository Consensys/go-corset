(column X :u8)
(column Y :u8)
;;
(module m1)
(column X :u8)
(column Y :u8)
(permute (A B) (+X +Y))
(vanish diag_ab (- (shift A 1) B))
