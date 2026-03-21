import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterLink } from '@angular/router';

@Component({
  selector: 'app-movie-seats',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './movie-seats.component.html',
  styleUrl: './movie-seats.component.scss'
})
export class MovieSeatsComponent implements OnInit {
  movieId = signal('');
  theaterId = signal<string | null>(null);
  showTime = signal<string | null>(null);

  constructor(private route: ActivatedRoute) {}

  ngOnInit() {
    this.route.paramMap.subscribe(pm => {
      this.movieId.set(pm.get('movieId') ?? '');
    });
    this.route.queryParamMap.subscribe(q => {
      this.theaterId.set(q.get('theaterId'));
      this.showTime.set(q.get('showTime'));
    });
  }
}
